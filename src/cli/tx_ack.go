package cli

import (
	// "fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"

	gcli "github.com/urfave/cli"

	messages "github.com/skycoin/hardware-wallet-protob/go"

	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func toUnit32Array(array *[]string) ([]uint32, error) {
	addressN := make([]uint32, len(*array))
	for index2, number := range *array {
		parseResult, err := strconv.ParseUint(number, 10, 32)
		if err != nil {
			return nil, err
		}
		addressN[index2] = uint32(parseResult)
	}
	return addressN, nil
}

func txAckCmd() gcli.Command {
	name := "txAck"
	return gcli.Command{
		Name:        name,
		Usage:       "Send a set of transactions or inputs requested by hardware wallet.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.IntFlag{
				Name:  "version",
				Usage: "Transaction version.",
			},
			gcli.IntFlag{
				Name:  "lockTime",
				Usage: "Lock time of transaction",
			},
			gcli.StringSliceFlag{
				Name:  "input",
				Usage: "Item with next structure: \"[address_n] [hashIn]\" .\n\t [address_n]: list of numbers separated for comma and represents BIP-32 path to derive the key from master node.\n\t [hashIn]: input hash.",
			},
			gcli.StringSliceFlag{
				Name:  "output",
				Usage: "Item with nest structure: \"[address] [address_n] [coins] [hours]\".\n\t [address]: target coin address in Base58 encoding. \n\t [address_n]: list of number separated for comma and represents BIP-32 path to derive the key from master node, has higher priority. \n\t [coins]: amount of transaction output.\n\t [hours]: accumulated hours.",
			},
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			version := uint32(c.Int("version"))
			lockTime := uint32(c.Int("lockTime"))
			inputsString := c.StringSlice("input")
			outputsString := c.StringSlice("output")

			device := skyWallet.NewDevice(skyWallet.DeviceTypeFromString(c.String("deviceType")))
			log.Info("Start Action on txAck function")
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
			txType := &messages.TxAck_TransactionType{
				Version:  &version,
				LockTime: &lockTime,
				Inputs:   nil,
				Outputs:  nil,
			}
			// Building txType
			if len(inputsString) != 0 {
				if len(outputsString) != 0 {
					log.Error("Inputs and outputs cannot be 0 in the same TxAck.")
					return
				}
				// Building Inputs
				Inputs := make([]*messages.TxAck_TransactionType_TxInputType, len(inputsString))
				for index, inputS := range inputsString {
					input := strings.Split(inputS, " ")
					/*
					* input[0] --> address_n
					* input[1] --> hashIn
					 */
					if len(input) != 2 {
						log.Error("Bad Input sintax: ", inputS)
						return
					}

					// Building address_n
					addressNString := strings.Split(input[0], ",")
					addressN, err := toUnit32Array(&addressNString)
					if err != nil {
						log.Error(err)
						return
					}

					// Buildeing hashIn
					hashIn := []byte(input[1])

					finalInput := &messages.TxAck_TransactionType_TxInputType{
						AddressN: addressN,
						HashIn:   hashIn,
					}
					Inputs[index] = finalInput
				}
				txType.Inputs = Inputs
			} else if len(outputsString) != 0 {
				// Building outputs
				Outputs := make([]*messages.TxAck_TransactionType_TxOutputType, len(outputsString))
				for index, outputS := range outputsString {
					output := strings.Split(outputS, " ")
					/*
					* output[0] --> address
					* output[1] --> address_n
					* output[2] --> coins
					* output[3] --> hours
					 */
					if len(output) != 4 {
						log.Error("Bad Output sintax: ", outputS)
					}

					// Building address
					address := output[0]

					// Building address_n
					addressNString := strings.Split(output[1], ",")
					addressN, err := toUnit32Array(&addressNString)
					if err != nil {
						log.Error(err)
						return
					}

					// Build coins
					coins, err := strconv.ParseUint(output[2], 10, 64)
					if err != nil {
						log.Error(err.Error())
						return
					}

					// Build hours
					hours, err := strconv.ParseUint(output[3], 10, 64)
					if err != nil {
						log.Error(err.Error())
						return
					}

					finalOutput := &messages.TxAck_TransactionType_TxOutputType{
						Address:  &address,
						AddressN: addressN,
						Coins:    &coins,
						Hours:    &hours,
					}
					Outputs[index] = finalOutput
				}
				txType.Outputs = Outputs
			}
			msg, err := device.TxAck(txType)
			if err != nil {
				log.Error(err.Error())
				return
			}

			switch msg.Kind {
			case uint16(messages.MessageType_MessageType_TxRequest):
				txRequest := &messages.TxRequest{}
				err := proto.Unmarshal(msg.Data, txRequest)
				if err != nil {
					log.Error(err.Error())
					return
				}
				println(txRequest.String())
			default:
				println("Unexpected response type\n")
			}
		},
	}
}
