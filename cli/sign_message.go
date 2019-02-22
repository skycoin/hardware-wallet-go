package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func signMessageCmd() gcli.Command {
	name := "signMessage"
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
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		OnUsageError: onCommandUsageError(name),
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

			addressN := c.Int("addressN")
			message := c.String("message")
			var signature string
			kind, data := deviceWallet.DeviceSignMessage(deviceType, addressN, message)
			for kind != uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) && kind != uint16(messages.MessageType_MessageType_Failure) {
				if kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
					var pinEnc string
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					kind, data = deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
					continue
				}

				if kind == uint16(messages.MessageType_MessageType_PassphraseRequest) {
					var passphrase string
					fmt.Printf("Input passphrase: ")
					fmt.Scanln(&passphrase)
					kind, data = deviceWallet.DevicePassphraseAck(deviceType, passphrase)
					continue
				}
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
