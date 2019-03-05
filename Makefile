all: build

build:
	go build ./...

dep:
	dep ensure
	# Ensure sources for protoc-gen-go and protobuf/proto are in sync
	dep ensure -add github.com/gogo/protobuf/protoc-gen-gofast

test:
	go test github.com/skycoin/hardware-wallet-go/device-wallet/

proto:
	make -C device-wallet/messages build-go GO_VENDOR_DIR=$$PWD/vendor

clean:
	make -C device-wallet/messages clean-go GO_VENDOR_DIR=$$PWD/vendor

lint:
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --skip-dirs=device-wallet/usb -E goimports -E golint -E varcheck -E unparam -E deadcode -E structcheck ./...

check: lint

format:
	goimports -w -local github.com/skycoin/hardware-wallet-go ./cli
	goimports -w -local github.com/skycoin/hardware-wallet-go ./device-wallet
