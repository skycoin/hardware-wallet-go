.DEFAULT_GOAL := help
.PHONY: all build dep vendor_proto mocks test proto clean lint check format

all: build

build: ## Build project
	cd cmd/cli && ./install.sh

dep: vendor_proto ## Ensure package dependencies are up to date
	dep ensure
	# Ensure sources for protoc-gen-go and protobuf/proto are in sync
	dep ensure -add github.com/gogo/protobuf/protoc-gen-gofast

vendor_proto: proto
	mkdir -p vendor/github.com/google/protobuf
	cp -r -p src/device-wallet/messages/go/google/protobuf/descriptor.pb.go vendor/github.com/google/protobuf

mocks: ## Create all mock files for unit tests
	echo "Generating mock files"
	mockery -all -dir ./interfaces -output ./test/mocks

test_unit: mocks ## Run unit tests
	go test -v github.com/skycoin/hardware-wallet-go/src/device-wallet

test_integration: ## Run integration tests
	go test -v github.com/skycoin/hardware-wallet-go/test/integration

test: test_unit test_integration ## Run all tests

proto: ## Generate protocol buffer classes for communicating with hardware wallet
	make -C src/device-wallet/messages build-go

clean: ## Delete temporary build files
	make -C src/device-wallet/messages clean-go
	rm -r vendor/github.com/google

lint: ## Check source code style
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --skip-dirs=src/device-wallet/usb test/mocks -E goimports -E golint -E varcheck -E unparam -E deadcode -E structcheck ./...

check: lint test ## Perform self-tests

format: ## Check and fix style
	goimports -w -local github.com/skycoin/hardware-wallet-go ./cmd
	goimports -w -local github.com/skycoin/hardware-wallet-go ./src

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
