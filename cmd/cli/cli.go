package main

import (
	"fmt"
	"os"

	"github.com/skycoin/hardware-wallet-go/src/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
