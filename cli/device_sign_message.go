package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func deviceSignMessageCmd() gcli.Command {
	name := "deviceSignMessage"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the device to sign a message using the secret key at given index.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.IntFlag{
				Name:  "addressN",
				Value: 0,
				Usage: "Index of the address that will issue the signature. Assume 0 if not set.",
			},
			gcli.StringFlag{
				Name:  "message",
				Usage: "The message that the signature claims to be signing.",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			addressN := c.Int("addressN")
			message := c.String("message")
			var signature string
			kind, data := deviceWallet.DeviceSignMessage(deviceWallet.DeviceTypeUsb, addressN, message)
			if kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				var pinEnc string
				fmt.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				kind, data = deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeUsb, pinEnc)
			}

			if kind == uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) {
				kind, signature = deviceWallet.DecodeResponseSkycoinSignMessage(kind, data)
				fmt.Printf("Success %d! the signature is: %s\n", kind, signature)
			} else {
				fmt.Printf("Failed with message: %s\n", deviceWallet.DecodeFailMsg(kind, data))
				return
			}
		},
	}
}
