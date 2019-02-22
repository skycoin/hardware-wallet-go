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
	protoc -I device-wallet/messages/  --gogofast_out=device-wallet/messages/ device-wallet/messages/messages.proto device-wallet/messages/types.proto device-wallet/messages/descriptor.proto

lint:
	golangci-lint run --no-config  --deadline=3m --concurrency=2 --skip-dirs=device-wallet/usb -E goimports -E golint -E varcheck -E unparam -E deadcode -E structcheck ./...

check: lint

format:
	goimports -w ./*.go device-wallet/
