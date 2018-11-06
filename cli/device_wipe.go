package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func deviceWipeCmd() gcli.Command {
	name := "deviceWipe"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the device to wipe clean all the configuration it contains.",
		Description: "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.WipeDevice(deviceWallet.DeviceTypeUsb)
		},
	}
}