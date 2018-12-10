package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorCancelCmd() gcli.Command {
	name := "emulatorCancel"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the emulator to cancel the ongoing procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.DeviceCancel(deviceWallet.DeviceTypeEmulator)
		},
	}
}
