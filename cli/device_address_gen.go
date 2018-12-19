package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func deviceAddressGenCmd() gcli.Command {
	name := "deviceAddressGen"
	return gcli.Command{
		Name:        name,
		Usage:       "Generate skycoin addresses using the firmware",
		Description: "",
		Flags: []gcli.Flag{
			gcli.IntFlag{
				Name:  "addressN",
				Value: 1,
				Usage: "Number of addresses to generate. Assume 1 if not set.",
			},
			gcli.IntFlag{
				Name:  "startIndex",
				Value: 0,
				Usage: "Index where deterministic key generation will start from. Assume 0 if not set.",
			},
			gcli.BoolFlag{
				Name:  "confirmAddress",
				Usage: "If requesting one address it will be sent only if user confirms operation by pressing device's button.",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			addressN := c.Int("addressN")
			startIndex := c.Int("startIndex")
			confirmAddress := c.Bool("confirmAddress")
			var data []byte
			var pinEnc string
			kind, data := deviceWallet.DeviceAddressGen(deviceWallet.DeviceTypeUsb, addressN, startIndex, confirmAddress)
			for kind != uint16(messages.MessageType_MessageType_ResponseSkycoinAddress) && kind != uint16(messages.MessageType_MessageType_Failure) {

				if kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					kind, data = deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeUsb, pinEnc)
					continue
				}
				if kind == uint16(messages.MessageType_MessageType_PassphraseRequest) {
					var passphrase string
					fmt.Printf("Input passphrase: ")
					fmt.Scanln(&passphrase)
					kind, data = deviceWallet.DevicePassphraseAck(deviceWallet.DeviceTypeUsb, passphrase)
					continue
				}

				if kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg := deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeUsb)
					kind, data = msg.Kind, msg.Data
					continue
				}
			}

			if kind == uint16(messages.MessageType_MessageType_ResponseSkycoinAddress) {
				_, addresses := deviceWallet.DecodeResponseSkycoinAddress(kind, data)
				fmt.Println(addresses)
			} else {
				fmt.Println(deviceWallet.DecodeFailMsg(kind, data))
			}
		},
	}
}
