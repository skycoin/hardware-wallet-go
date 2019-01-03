all: build

build:
	go build ./...

deps:
	dep ensure

test:
	go test github.com/skycoin/hardware-wallet-go/device-wallet/
