// Copyright 2015 The MFA Authors
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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/MFAChain/mfachain/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("mfa/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("mfa/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("mfa/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("mfa/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("mfa/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("mfa/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("mfa/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("mfa/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("mfa/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("mfa/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("mfa/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("mfa/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("mfa/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("mfa/downloader/states/drop", nil)
)
