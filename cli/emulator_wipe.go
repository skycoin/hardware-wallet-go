package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorWipeCmd() gcli.Command {
	name := "emulatorWipe"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the emulator to wipe clean all the configuration it contains.",
		Description: "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.WipeDevice(deviceWallet.DeviceTypeEmulator)
		},
	}
}