package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func emulatorBackupCmd() gcli.Command {
	name := "emulatorBackup"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the emulator to perform the seed backup procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			deviceWallet.BackupDevice(deviceWallet.DeviceTypeEmulator)
		},
	}
}
