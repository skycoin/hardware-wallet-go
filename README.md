# Go bindings and CLI tool for the Skycoin hardware wallet

[![Build Status](https://travis-ci.com/skycoin/hardware-wallet-go.svg?branch=master)](https://travis-ci.com/skycoin/hardware-wallet-go)

## Installation

### Install golang

    https://github.com/golang/go/wiki/Ubuntu

## Usage

### Download source code

```bash
$ go get github.com/skycoin/hardware-wallet-go
```

### Dependancies management

This project uses dep [dependancy manager](https://github.com/golang/dep).

Don't modify anything under vendor/ directory without using [dep commands](https://github.com/golang/dep/blob/master/docs/Gopkg.toml.md).

Download dependencies using command:

```bash
$ make dep
```

### Generate protobuf files

#### Init proto submboule
```bash
$ git submodule init 
$ git submodule update
```

#### Generate go files
```bash
$ make vendor_proto
```

### Run

```bash
$ go run cmd/cli/cli.go
```

See also [CLI README](https://github.com/skycoin/hardware-wallet-go/blob/master/cmd/cli/README.md) for information about the Command Line Interface.

## Wiki

More information in [the wiki](https://github.com/skycoin/hardware-wallet-go/wiki)
