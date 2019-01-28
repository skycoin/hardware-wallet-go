package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"
	// deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	// "github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func emulatorTransactionSignCmd() gcli.Command {
	name := "emulatorTransactionSign"
	return gcli.Command{
		Name:        name,
		Usage:       "Ask the device to sign a transaction using the provided information.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.StringSliceFlag{
				Name:  "inputHash",
				Usage: "Hash of the Input of the transaction we expect the device to sign",
			},
			gcli.IntSliceFlag{
				Name:  "inputIndex",
				Usage: "Index of the input in the wallet",
			},
			gcli.StringSliceFlag{
				Name:  "outputAddress",
				Usage: "Addresses of the output for the transaction",
			},
			gcli.IntSliceFlag{
				Name:  "coin",
				Usage: "Amount of coins",
			},
			gcli.IntSliceFlag{
				Name:  "hour",
				Usage: "Number of hours",
			},
			gcli.IntSliceFlag{
				Name:  "addressIndex",
				Usage: "If the address is a return address tell its index in the wallet",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			inputs := c.StringSlice("inputHash")
			inputIndex := c.IntSlice("inputIndex")
			outputs := c.StringSlice("outputAddress")
			coin := c.IntSlice("coin")
			hour := c.IntSlice("hour")
			addressIndex := c.IntSlice("addressIndex")

			fmt.Println(inputs, inputIndex)
			fmt.Println(outputs, coin, hour, addressIndex)
			/*
				kind, data := deviceWallet.DeviceTransactionSign(deviceWallet.DeviceTypeEmulator, inputs, outputs)
				for {
					switch kind {
					case uint16(messages.MessageType_MessageType_ResponseTransactionSign):
						kind, signatures := deviceWallet.DecodeResponseTransactionSign(kind, data)
						fmt.Println(signatures)
						return
					case uint16(messages.MessageType_MessageType_Success):
						fmt.Println("Should end with ResponseTransactionSign request")
						return
					case uint16(messages.MessageType_MessageType_ButtonRequest):
						msg := deviceWallet.DeviceButtonAck(deviceWallet.DeviceTypeUsb)
						kind, data = msg.Kind, msg.Data
					case uint16(messages.MessageType_MessageType_PassphraseRequest):
						var passphrase string
						fmt.Printf("Input passphrase: ")
						fmt.Scanln(&passphrase)
						kind, data = deviceWallet.DevicePassphraseAck(deviceWallet.DeviceTypeEmulator, passphrase)
					case uint16(messages.MessageType_MessageType_PinMatrixRequest):
						var pinEnc string
						fmt.Printf("PinMatrixRequest response: ")
						fmt.Scanln(&pinEnc)
						kind, data = deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeEmulator, pinEnc)
					case uint16(messages.MessageType_MessageType_Failure):
					default:
						fmt.Printf("Failed with message: %s\n", deviceWallet.DecodeFailMsg(kind, data))
						return
					}
				}
			*/
		},
	}
}
