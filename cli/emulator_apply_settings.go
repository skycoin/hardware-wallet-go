package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
	"github.com/skycoin/hardware-wallet-go/device-wallet/wire"
)

func emulatorApplySettingsCmd() gcli.Command {
	name := "emulatorApplySettings"
	return gcli.Command{
		Name:        name,
		Usage:       "Apply settings.",
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
			msg := deviceWallet.DeviceApplySettings(deviceWallet.DeviceTypeEmulator, passphrase)
			for msg.Kind != uint16(messages.MessageType_MessageType_Failure) && msg.Kind != uint16(messages.MessageType_MessageType_Success) {

				if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg = deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeEmulator)
					continue
				}

				if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
					var pinEnc string
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					kind, data := deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeEmulator, pinEnc)
					msg = wire.Message{
						Kind: kind,
						Data: data,
					}
					continue
				}
			}
			if msg.Kind == uint16(messages.MessageType_MessageType_Failure) {
				failMsg := deviceWallet.DecodeFailMsg(msg.Kind, msg.Data)
				fmt.Println("Failed with code: ", failMsg)
				return
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_Success) {
				successMsg := deviceWallet.DecodeSuccessMsg(msg.Kind, msg.Data)
				fmt.Println("Success with code: ", successMsg)
				return
			}
		},
	}
}
