package cli

import (
	gcli "github.com/urfave/cli"

	"log"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func checkMessageSignatureCmd() gcli.Command {
	name := "checkMessageSignature"
	return gcli.Command{
		Name:        name,
		Usage:       "Check a message signature matches the given address.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.StringFlag{
				Name:  "message",
				Usage: "The message that the signature claims to be signing.",
			},
			gcli.StringFlag{
				Name:  "signature",
				Usage: "Signature of the message.",
			},
			gcli.StringFlag{
				Name:  "address",
				Usage: "Address that issued the signature.",
			},
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
				Value:  "USB",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			message := c.String("message")
			signature := c.String("signature")
			address := c.String("address")

			var deviceType deviceWallet.DeviceType
			switch c.String("deviceType") {
			case "USB":
				deviceType = deviceWallet.DeviceTypeUsb
			case "EMULATOR":
				deviceType = deviceWallet.DeviceTypeEmulator
			default:
				log.Println("No device detected")
				return
			}

			deviceWallet.DeviceCheckMessageSignature(deviceType, message, signature, address)
		},
	}
}
