package cli

import (
	"fmt"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
	gcli "github.com/urfave/cli"
)

func emulatorRecoveryCmd() gcli.Command {
	name := "emulatorRecovery"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the emulator to perform the seed recovery procedure.",
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
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			passphrase := c.Bool("usePassphrase")
			dryRun := c.Bool("dryRun")
			wordCount := uint32(c.Uint64("wordCount"))
			msg := deviceWallet.RecoveryDevice(deviceWallet.DeviceTypeEmulator, wordCount, passphrase, dryRun)
			for msg.Kind == uint16(messages.MessageType_MessageType_WordRequest) {
				var word string
				fmt.Printf("Word: ")
				fmt.Scanln(&word)
				msg = deviceWallet.DeviceWordAck(deviceWallet.DeviceTypeEmulator, word)
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				// Send ButtonAck
				msg = deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeEmulator)
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
