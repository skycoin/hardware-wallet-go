package cli

import (
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
			deviceWallet.DeviceGetVersion(deviceWallet.DeviceTypeEmulator)
		},
	}
}
