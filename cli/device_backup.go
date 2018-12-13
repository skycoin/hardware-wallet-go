package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"

	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func deviceBackupCmd() gcli.Command {
	name := "deviceBackup"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to perform the seed backup procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			msg := deviceWallet.BackupDevice(deviceWallet.DeviceTypeUsb)
			if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				var pinEnc string
				fmt.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				kind, _ := deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeUsb, pinEnc)
				for kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg = deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeUsb)
				}

			}
			fmt.Println(deviceWallet.DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
		},
	}
}
