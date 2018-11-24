package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorGetVersionCmd() gcli.Command {
	name := "emulatorGetVersion"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask firmware version.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			version := deviceWallet.DeviceGetVersion(deviceWallet.DeviceTypeEmulator)
			fmt.Printf("Firmware version is %s\n", version)
		},
	}
}
