package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func deviceBackupCmd() gcli.Command {
	name := "deviceBackup"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to perform the seed backup procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.BackupDevice(deviceWallet.DeviceTypeUsb)
		},
	}
}
