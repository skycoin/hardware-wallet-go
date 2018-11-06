package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func deviceGenerateMnemonicCmd() gcli.Command {
	name := "deviceGenerateMnemonic"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to generate a mnemonic and configure itself with it.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.DeviceGenerateMnemonic(deviceWallet.DeviceTypeUsb)
		},
	}
}
