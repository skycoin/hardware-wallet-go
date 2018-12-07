package cli

import (
	gcli "github.com/urfave/cli"

	emulatorWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorFeaturesCmd() gcli.Command {
	name := "emulatorFeatures"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the emulator Features.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			emulatorWallet.DeviceGetFeatures(emulatorWallet.DeviceTypeEmulator)
		},
	}
}
