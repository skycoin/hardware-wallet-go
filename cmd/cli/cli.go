package main

import (
	"fmt"
	"os"

	"github.com/skycoin/hardware-wallet-go/src/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
