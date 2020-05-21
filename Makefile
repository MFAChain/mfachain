# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: mfa android ios mfa-cross evm all test clean
.PHONY: mfa-linux mfa-linux-386 mfa-linux-amd64 mfa-linux-mips64 mfa-linux-mips64le
.PHONY: mfa-linux-arm mfa-linux-arm-5 mfa-linux-arm-6 mfa-linux-arm-7 mfa-linux-arm64
.PHONY: mfa-darwin mfa-darwin-386 mfa-darwin-amd64
.PHONY: mfa-windows mfa-windows-386 mfa-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

mfa:
	$(GORUN) build/ci.go install ./cmd/mfa
	@echo "Done building."
	@echo "Run \"$(GOBIN)/mfa\" to launch mfa."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/mfa.aar\" to use the library."

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Mfa.framework\" to use the library."

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

mfa-cross: mfa-linux mfa-darwin mfa-windows mfa-android mfa-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/mfa-*

mfa-linux: mfa-linux-386 mfa-linux-amd64 mfa-linux-arm mfa-linux-mips64 mfa-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-*

mfa-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/mfa
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep 386

mfa-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/mfa
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep amd64

mfa-linux-arm: mfa-linux-arm-5 mfa-linux-arm-6 mfa-linux-arm-7 mfa-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep arm

mfa-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/mfa
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep arm-5

mfa-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/mfa
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep arm-6

mfa-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/mfa
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep arm-7

mfa-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/mfa
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep arm64

mfa-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/mfa
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep mips

mfa-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/mfa
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep mipsle

mfa-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/mfa
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep mips64

mfa-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/mfa
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/mfa-linux-* | grep mips64le

mfa-darwin: mfa-darwin-386 mfa-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/mfa-darwin-*

mfa-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/mfa
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-darwin-* | grep 386

mfa-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/mfa
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-darwin-* | grep amd64

mfa-windows: mfa-windows-386 mfa-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/mfa-windows-*

mfa-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/mfa
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-windows-* | grep 386

mfa-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/mfa
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/mfa-windows-* | grep amd64
