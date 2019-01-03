all: build

build:
	go build ./...

deps:
	dep ensure

test:
	go test github.com/skycoin/hardware-wallet-go/device-wallet/

proto:
	protoc -I device-wallet/messages/  --go_out=device-wallet/messages/ device-wallet/messages/messages.proto device-wallet/messages/types.proto device-wallet/messages/descriptor.proto
