package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
	"github.com/skycoin/hardware-wallet-go/device-wallet/wire"
)

func emulatorBackupCmd() gcli.Command {
	name := "emulatorBackup"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the emulator to perform the seed backup procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			msg := deviceWallet.BackupDevice(deviceWallet.DeviceTypeEmulator)
			if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				var pinEnc string
				fmt.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				kind, data := deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeEmulator, pinEnc)
				msg = wire.Message{
					Kind: kind,
					Data: data,
				}
				for msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg = deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeEmulator, msg)
				}

			}
			fmt.Println(deviceWallet.DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
		},
	}
}
