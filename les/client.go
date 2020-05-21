// Copyright 2019 The MFA Authors
// This file is part of this library.
//
// This library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light MFA Subprotocol.
package les

import (
	"fmt"
	"time"

	"github.com/MFAChain/mfachain/accounts"
	"github.com/MFAChain/mfachain/accounts/abi/bind"
	"github.com/MFAChain/mfachain/common"
	"github.com/MFAChain/mfachain/common/hexutil"
	"github.com/MFAChain/mfachain/common/mclock"
	"github.com/MFAChain/mfachain/consensus"
	"github.com/MFAChain/mfachain/core"
	"github.com/MFAChain/mfachain/core/bloombits"
	"github.com/MFAChain/mfachain/core/rawdb"
	"github.com/MFAChain/mfachain/core/types"
	"github.com/MFAChain/mfachain/eth"
	"github.com/MFAChain/mfachain/eth/downloader"
	"github.com/MFAChain/mfachain/eth/filters"
	"github.com/MFAChain/mfachain/eth/gasprice"
	"github.com/MFAChain/mfachain/event"
	"github.com/MFAChain/mfachain/internal/ethapi"
	"github.com/MFAChain/mfachain/les/checkpointoracle"
	lpc "github.com/MFAChain/mfachain/les/lespay/client"
	"github.com/MFAChain/mfachain/light"
	"github.com/MFAChain/mfachain/log"
	"github.com/MFAChain/mfachain/node"
	"github.com/MFAChain/mfachain/p2p"
	"github.com/MFAChain/mfachain/p2p/enode"
	"github.com/MFAChain/mfachain/params"
	"github.com/MFAChain/mfachain/rpc"
)

type LightMFA struct {
	lesCommons

	peers        *serverPeerSet
	reqDist      *requestDistributor
	retriever    *retrieveManager
	odr          *LesOdr
	relay        *lesTxRelay
	handler      *clientHandler
	txPool       *light.TxPool
	blockchain   *light.LightChain
	serverPool   *serverPool
	valueTracker *lpc.ValueTracker

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend     *LesApiBackend
	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager
	netRPCService  *ethapi.PublicNetAPI
}

func New(ctx *node.ServiceContext, config *eth.Config) (*LightMFA, error) {
	chainDb, err := ctx.OpenDatabase("lightchaindata", config.DatabaseCache, config.DatabaseHandles, "mfa/db/chaindata/")
	if err != nil {
		return nil, err
	}
	lespayDb, err := ctx.OpenDatabase("lespay", 0, 0, "mfa/db/lespay")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newServerPeerSet()
	leth := &LightMFA{
		lesCommons: lesCommons{
			genesis:     genesisHash,
			config:      config,
			chainConfig: chainConfig,
			iConfig:     light.DefaultClientIndexerConfig,
			chainDb:     chainDb,
			closeCh:     make(chan struct{}),
		},
		peers:          peers,
		eventMux:       ctx.EventMux,
		reqDist:        newRequestDistributor(peers, &mclock.System{}),
		accountManager: ctx.AccountManager,
		engine:         eth.CreateConsensusEngine(ctx, chainConfig, &config.Ethash, nil, false, chainDb),
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   eth.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
		serverPool:     newServerPool(chainDb, config.UltraLightServers),
		valueTracker:   lpc.NewValueTracker(lespayDb, &mclock.System{}, requestList, time.Minute, 1/float64(time.Hour), 1/float64(time.Hour*100), 1/float64(time.Hour*1000)),
	}
	peers.subscribe((*vtSubscription)(leth.valueTracker))
	leth.retriever = newRetrieveManager(peers, leth.reqDist, leth.serverPool)
	leth.relay = newLesTxRelay(peers, leth.retriever)

	leth.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, leth.retriever)
	leth.chtIndexer = light.NewChtIndexer(chainDb, leth.odr, params.CHTFrequency, params.HelperTrieConfirmations)
	leth.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, leth.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	leth.odr.SetIndexers(leth.chtIndexer, leth.bloomTrieIndexer, leth.bloomIndexer)

	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if leth.blockchain, err = light.NewLightChain(leth.odr, leth.chainConfig, leth.engine, checkpoint); err != nil {
		return nil, err
	}
	leth.chainReader = leth.blockchain
	leth.txPool = light.NewTxPool(leth.chainConfig, leth.blockchain, leth.relay)

	// Set up checkpoint oracle.
	oracle := config.CheckpointOracle
	if oracle == nil {
		oracle = params.CheckpointOracles[genesisHash]
	}
	leth.oracle = checkpointoracle.New(oracle, leth.localCheckpoint)

	// Note: AddChildIndexer starts the update process for the child
	leth.bloomIndexer.AddChildIndexer(leth.bloomTrieIndexer)
	leth.chtIndexer.Start(leth.blockchain)
	leth.bloomIndexer.Start(leth.blockchain)

	leth.handler = newClientHandler(config.UltraLightServers, config.UltraLightFraction, checkpoint, leth)
	if leth.handler.ulc != nil {
		log.Warn("Ultra light client is enabled", "trustedNodes", len(leth.handler.ulc.keys), "minTrustedFraction", leth.handler.ulc.fraction)
		leth.blockchain.DisableCheckFreq()
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		leth.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	leth.ApiBackend = &LesApiBackend{ctx.ExtRPCEnabled(), leth, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	leth.ApiBackend.gpo = gasprice.NewOracle(leth.ApiBackend, gpoParams)

	return leth, nil
}

// vtSubscription implements serverPeerSubscriber
type vtSubscription lpc.ValueTracker

// registerPeer implements serverPeerSubscriber
func (v *vtSubscription) registerPeer(p *serverPeer) {
	vt := (*lpc.ValueTracker)(v)
	p.setValueTracker(vt, vt.Register(p.ID()))
	p.updateVtParams()
}

// unregisterPeer implements serverPeerSubscriber
func (v *vtSubscription) unregisterPeer(p *serverPeer) {
	vt := (*lpc.ValueTracker)(v)
	vt.Unregister(p.ID())
	p.setValueTracker(nil, nil)
}

type LightDummyAPI struct{}

// Etherbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Etherbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the MFA package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightMFA) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.ApiBackend)
	apis = append(apis, s.engine.APIs(s.BlockChain().HeaderChain())...)
	return append(apis, []rpc.API{
		{
			Namespace: "mfa",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "mfa",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.handler.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "mfa",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "les",
			Version:   "1.0",
			Service:   NewPrivateLightAPI(&s.lesCommons),
			Public:    false,
		}, {
			Namespace: "lespay",
			Version:   "1.0",
			Service:   lpc.NewPrivateClientAPI(s.valueTracker),
			Public:    false,
		},
	}...)
}

func (s *LightMFA) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightMFA) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightMFA) TxPool() *light.TxPool              { return s.txPool }
func (s *LightMFA) Engine() consensus.Engine           { return s.engine }
func (s *LightMFA) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightMFA) Downloader() *downloader.Downloader { return s.handler.downloader }
func (s *LightMFA) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightMFA) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions, s.handler.runPeer, func(id enode.ID) interface{} {
		if p := s.peers.peer(peerIdToString(id)); p != nil {
			return p.Info()
		}
		return nil
	})
}

// Start implements node.Service, starting all internal goroutines needed by the
// light MFA protocol implementation.
func (s *LightMFA) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")

	// Start bloom request workers.
	s.wg.Add(bloomServiceThreads)
	s.startBloomHandlers(params.BloomBitsBlocksClient)

	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.config.NetworkId)

	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// MFA protocol.
func (s *LightMFA) Stop() error {
	close(s.closeCh)
	s.peers.close()
	s.reqDist.close()
	s.odr.Stop()
	s.relay.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.handler.stop()
	s.txPool.Stop()
	s.engine.Close()
	s.eventMux.Stop()
	s.serverPool.stop()
	s.valueTracker.Stop()
	s.chainDb.Close()
	s.wg.Wait()
	log.Info("Light MFA stopped")
	return nil
}

// SetClient sets the rpc client and binds the registrar contract.
func (s *LightMFA) SetContractBackend(backend bind.ContractBackend) {
	if s.oracle == nil {
		return
	}
	s.oracle.Start(backend)
}
