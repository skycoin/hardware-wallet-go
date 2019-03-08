all: build

build:
	cd cmd/cli && ./install.sh

dep: vendor_proto
	dep ensure
	# Ensure sources for protoc-gen-go and protobuf/proto are in sync
	dep ensure -add github.com/gogo/protobuf/protoc-gen-gofast

vendor_proto: proto
	mkdir -p vendor/github.com/google/protobuf
	cp -r -p src/device-wallet/messages/go/google/protobuf/descriptor.pb.go vendor/github.com/google/protobuf


test:
	go test github.com/skycoin/hardware-wallet-go/src/device-wallet

proto:
	make -C src/device-wallet/messages build-go

clean:
	make -C src/device-wallet/messages clean-go
	rm -r vendor/github.com/google

lint:
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --skip-dirs=src/device-wallet/usb -E goimports -E golint -E varcheck -E unparam -E deadcode -E structcheck ./...

check: lint

format:
	goimports -w -local github.com/skycoin/hardware-wallet-go ./cmd
	goimports -w -local github.com/skycoin/hardware-wallet-go ./src
