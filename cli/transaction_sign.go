package cli

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
	gcli "github.com/urfave/cli"
)

func transactionSignCmd() gcli.Command {
	name := "transactionSign"
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
			gcli.Int64SliceFlag{
				Name:  "coin",
				Usage: "Amount of coins",
			},
			gcli.Int64SliceFlag{
				Name:  "hour",
				Usage: "Number of hours",
			},
			gcli.IntSliceFlag{
				Name:  "addressIndex",
				Usage: "If the address is a return address tell its index in the wallet",
			},
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			inputs := c.StringSlice("inputHash")
			inputIndex := c.IntSlice("inputIndex")
			outputs := c.StringSlice("outputAddress")
			coins := c.Int64Slice("coin")
			hours := c.Int64Slice("hour")
			addressIndex := c.IntSlice("addressIndex")

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

			fmt.Println(inputs, inputIndex)
			if len(inputs) != len(inputIndex) {
				fmt.Println("Every given input hash should have the an inputIndex")
				return
			}
			if len(outputs) != len(coins) || len(outputs) != len(hours) {
				fmt.Println("Every given output should have a coin and hour value")
				return
			}
			fmt.Println(outputs, coins, hours, addressIndex)
			var transactionInputs []*messages.SkycoinTransactionInput
			var transactionOutputs []*messages.SkycoinTransactionOutput
			for i, input := range inputs {
				var transactionInput messages.SkycoinTransactionInput
				transactionInput.HashIn = proto.String(input)
				transactionInput.Index = proto.Uint32(uint32(inputIndex[i]))
				transactionInputs = append(transactionInputs, &transactionInput)
			}
			for i, output := range outputs {
				var transactionOutput messages.SkycoinTransactionOutput
				transactionOutput.Address = proto.String(output)
				transactionOutput.Coin = proto.Uint64(uint64(coins[i]))
				transactionOutput.Hour = proto.Uint64(uint64(hours[i]))
				if i < len(addressIndex) {
					transactionOutput.AddressIndex = proto.Uint32(uint32(addressIndex[i]))
				}
				transactionOutputs = append(transactionOutputs, &transactionOutput)
			}
			kind, data := deviceWallet.DeviceTransactionSign(deviceType, transactionInputs, transactionOutputs)
			for {
				switch kind {
				case uint16(messages.MessageType_MessageType_ResponseTransactionSign):
					_, signatures := deviceWallet.DecodeResponseTransactionSign(kind, data)
					fmt.Println(signatures)
					return
				case uint16(messages.MessageType_MessageType_Success):
					fmt.Println("Should end with ResponseTransactionSign request")
					return
				case uint16(messages.MessageType_MessageType_ButtonRequest):
					msg := deviceWallet.DeviceButtonAck(deviceType)
					kind, data = msg.Kind, msg.Data
				case uint16(messages.MessageType_MessageType_PassphraseRequest):
					var passphrase string
					fmt.Printf("Input passphrase: ")
					fmt.Scanln(&passphrase)
					kind, data = deviceWallet.DevicePassphraseAck(deviceType, passphrase)
				case uint16(messages.MessageType_MessageType_PinMatrixRequest):
					var pinEnc string
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					kind, data = deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
				case uint16(messages.MessageType_MessageType_Failure):
					fmt.Printf("Failed with message: %s\n", deviceWallet.DecodeFailMsg(kind, data))
					return
				default:
					fmt.Printf("Failed with message: %s\n", deviceWallet.DecodeFailMsg(kind, data))
					return
				}
			}
		},
	}
}
