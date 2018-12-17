package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorGenerateMnemonicCmd() gcli.Command {
	name := "emulatorGenerateMnemonic"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the device to generate a mnemonic and configure itself with it.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.BoolFlag{
				Name:  "usePassphrase",
				Usage: "Configure a passphrase",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			passphrase := c.Bool("usePassphrase")
			deviceWallet.DeviceGenerateMnemonic(deviceWallet.DeviceTypeEmulator, passphrase)
		},
	}
}
