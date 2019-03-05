package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func recoveryCmd() gcli.Command {
	name := "recovery"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the device to perform the seed recovery procedure.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.BoolFlag{
				Name:  "usePassphrase",
				Usage: "Configure a passphrase",
			},
			gcli.BoolFlag{
				Name:  "dryRun",
				Usage: "perform dry-run recovery workflow (for safe mnemonic validation)",
			},
			gcli.IntFlag{
				Name:  "wordCount",
				Usage: "Use a specific (12 | 24) number of words for the Mnemonic recovery",
				Value: 12,
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

			passphrase := c.Bool("usePassphrase")
			dryRun := c.Bool("dryRun")
			wordCount := uint32(c.Uint64("wordCount"))
			msg, err := deviceWallet.RecoveryDevice(deviceType, wordCount, passphrase, dryRun)
			if err != nil {
				log.Error(err)
				return
			}

			for msg.Kind == uint16(messages.MessageType_MessageType_WordRequest) {
				var word string
				fmt.Printf("Word: ")
				fmt.Scanln(&word)
				msg, err = deviceWallet.DeviceWordAck(deviceType, word)
				if err != nil {
					log.Error(err)
					return
				}
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				// Send ButtonAck
				msg, err = deviceWallet.DeviceButtonAck(deviceType)
				if err != nil {
					log.Error(err)
					return
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
