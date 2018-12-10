package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func deviceCancelCmd() gcli.Command {
	name := "deviceCancel"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to cancel the ongoing procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.DeviceCancel(deviceWallet.DeviceTypeUsb)
		},
	}
}
