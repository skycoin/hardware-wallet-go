all: build

build:
	cd cmd/cli && ./install.sh

dep:
	dep ensure
	# Ensure sources for protoc-gen-go and protobuf/proto are in sync
	dep ensure -add github.com/gogo/protobuf/protoc-gen-gofast

test:
	go test github.com/skycoin/hardware-wallet-go/src

proto:
	protoc -I src/device-wallet/messages/  --gogofast_out=src/device-wallet/messages/ src/device-wallet/messages/messages.proto src/device-wallet/messages/types.proto src/device-wallet/messages/descriptor.proto

lint:
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --skip-dirs=src/device-wallet/usb -E goimports -E golint -E varcheck -E unparam -E deadcode -E structcheck ./...

check: lint

format:
	goimports -w -local github.com/skycoin/hardware-wallet-go ./cmd
	goimports -w -local github.com/skycoin/hardware-wallet-go ./src
