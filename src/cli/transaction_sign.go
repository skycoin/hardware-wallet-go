package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gogo/protobuf/proto"

	gcli "github.com/urfave/cli"

	messages "github.com/SkycoinProject/hardware-wallet-protob/go"

	skyWallet "github.com/SkycoinProject/hardware-wallet-go/src/skywallet"
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
			gcli.StringSliceFlag{
				Name:  "prevHash",
				Usage: "Hash of the previous transaction we expect the device to sign",
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
			gcli.StringFlag{
				Name:   "coinType",
				Value:  "SKY",
				Usage:  "Coin type to use on hardware-wallet.",
				EnvVar: "COIN_TYPE",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			inputs := c.StringSlice("inputHash")
			prevHash := c.StringSlice("prevHash")
			inputIndex := c.IntSlice("inputIndex")
			outputs := c.StringSlice("outputAddress")
			coins := c.Int64Slice("coin")
			hours := c.Int64Slice("hour")
			addressIndex := c.IntSlice("addressIndex")
			coinType, err := skyWallet.CoinTypeFromString(c.String("coinType"))
			if err != nil {
				log.Error(err)
				return
			}
			if coinType != skyWallet.SkycoinCoinType && len(inputs) > 0 {
				log.Error(fmt.Errorf("coin type %s doesn't need input hash", coinType))
				return
			}

			if coinType != skyWallet.BitcoinCoinType && len(prevHash) > 0 {
				log.Error(fmt.Errorf("coin type %s doesn't need previous hash", coinType))
				return
			}

			device := skyWallet.NewDevice(skyWallet.DeviceTypeFromString(c.String("deviceType")))
			if device == nil {
				return
			}
			defer device.Close()

			if os.Getenv("AUTO_PRESS_BUTTONS") == "1" && device.Driver.DeviceType() == skyWallet.DeviceTypeEmulator && runtime.GOOS == "linux" {
				err := device.SetAutoPressButton(true, skyWallet.ButtonRight)
				if err != nil {
					log.Error(err)
					return
				}
			}

			if len(inputs) != len(inputIndex) {
				fmt.Println("Every given input hash should have the an inputIndex")
				return
			}
			if len(outputs) != len(coins) {
				fmt.Println("Every given output should have a coin value")
				return
			}

			switch coinType {
			case skyWallet.SkycoinCoinType:
				err = transactionSkycoinSign(device, inputs, outputs, coins, hours, inputIndex, addressIndex)
				if err != nil {
					log.Error(err)
				}
			default:
				log.Error(fmt.Errorf("TransactionSign is not implemented for %s yet", coinType))
			}
		},
	}
}

func transactionSkycoinSign(device *skyWallet.Device, inputs, outputs []string, coins, hours []int64, inputIndex, addressIndex []int) error {
	if len(outputs) != len(hours) {
		return fmt.Errorf("Every given output should have a coin value")
	}

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

	msg, err := device.TransactionSign(transactionInputs, transactionOutputs)
	if err != nil {
		return err
	}

	for {
		switch msg.Kind {
		case uint16(messages.MessageType_MessageType_ResponseTransactionSign):
			signatures, err := skyWallet.DecodeResponseTransactionSign(msg)
			if err != nil {
				return err
			}
			fmt.Println(signatures)
			return nil
		case uint16(messages.MessageType_MessageType_Success):
			fmt.Println("Should end with ResponseTransactionSign request")
			return nil
		case uint16(messages.MessageType_MessageType_ButtonRequest):
			msg, err = device.ButtonAck()
			if err != nil {
				return err
			}
		case uint16(messages.MessageType_MessageType_PassphraseRequest):
			var passphrase string
			fmt.Printf("Input passphrase: ")
			fmt.Scanln(&passphrase)
			msg, err = device.PassphraseAck(passphrase)
			if err != nil {
				return err
			}
		case uint16(messages.MessageType_MessageType_PinMatrixRequest):
			var pinEnc string
			fmt.Printf("PinMatrixRequest response: ")
			fmt.Scanln(&pinEnc)
			msg, err = device.PinMatrixAck(pinEnc)
			if err != nil {
				return err
			}
		case uint16(messages.MessageType_MessageType_Failure):
			failMsg, err := skyWallet.DecodeFailMsg(msg)
			if err != nil {
				fmt.Printf("Failed with message: %s\n", failMsg)
				return err
			}

		default:
			return fmt.Errorf("received unexpected message type: %s", messages.MessageType(msg.Kind))
		}
	}
}
