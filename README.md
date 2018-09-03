# Go bindings and CLI tool for the Skycoin hardware wallet

[![Build Status](https://travis-ci.com/skycoin/hardware-wallet-go.svg?branch=master)](https://travis-ci.com/skycoin/hardware-wallet-go)

## Installation

### Install golang

    https://github.com/golang/go/wiki/Ubuntu

## Usage

### Download source code
    
    go get github.com/skycoin/hardware-wallet-go

### Dependancies management

This project uses dep [dependancy manager](https://github.com/golang/dep).

Don't modify anything under vendor/ directory without using [dep commands](https://github.com/golang/dep/blob/master/docs/Gopkg.toml.md).

### Generate protobuf files

Only once each time the messages change:

    cd device-wallet/
    protoc -I ./protob  --go_out=./protob protob/messages.proto protob/types.proto protob/descriptor.proto 

### Run

    go test -run TestMain

## Wiki

More information in [the wiki](https://github.com/skycoin/hardware-wallet-go/wiki)
