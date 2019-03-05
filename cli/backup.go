package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"

	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func backupCmd() gcli.Command {
	name := "backup"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to perform the seed backup procedure.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Flags: []gcli.Flag{
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		Action: func(c *gcli.Context) {
			var deviceType deviceWallet.DeviceType
			switch c.String("deviceType") {
			case "USB":
				deviceType = deviceWallet.DeviceTypeUsb
			case "EMULATOR":
				deviceType = deviceWallet.DeviceTypeEmulator
			default:
				log.Error("device type not set")
				return
			}

			msg, err := deviceWallet.BackupDevice(deviceType)
			if err != nil {
				log.Error(err)
				return
			}
			if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				var pinEnc string
				fmt.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				msg, err := deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
				if err != nil {
					log.Error(err)
					return
				}

				// TODO: can DeviceButtonAck return MessageType_MessageType_ButtonRequest? figure out
				for msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg, err = deviceWallet.DeviceButtonAck(deviceType)
					if err != nil {
						log.Error(err)
						return
					}
				}
			}

			responseMsg, err := deviceWallet.DecodeSuccessOrFailMsg(msg)
			if err != nil {
				log.Error(err)
				return
			}

			fmt.Println(responseMsg)
		},
	}
}
