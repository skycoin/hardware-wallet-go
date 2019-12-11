package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/SkycoinProject/hardware-wallet-go/src/skywallet/wire"

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
			coinName := "Skycoin"
			version := 1
			lockTime := 0
			txHash := "dkdji9e2oidhash"

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
			if len(outputs) != len(coins) || len(outputs) != len(hours) {
				fmt.Println("Every given output should have a coin and hour value")
				return
			}

			if len(inputs) > 7 || len(outputs) > 7 {
				state := 0
				index := 0

				msg, err := device.SignTx(len(outputs), len(inputs), coinName, version, lockTime, txHash)

				for {
					if err != nil {
						log.Error(err)
						return
					}
					switch msg.Kind {
					case uint16(messages.MessageType_MessageType_TxRequest):
						txRequest := &messages.TxRequest{}
						err = proto.Unmarshal(msg.Data, txRequest)
						if err != nil {
							log.Error(err)
							return
						}
						switch *txRequest.RequestType {
						case messages.TxRequest_TXINPUT:
							if state == 0 { // Sending Inputs for InnerHash
								msg, err = sendInputs(device, &inputs, &inputIndex, version, lockTime, &index, &state)
							} else if state == 2 { // Sending Inputs for Signatures
								err = printSignatures(&msg)
								if err != nil {
									log.Error(err)
									return
								}
								msg, err = sendInputs(device, &inputs, &inputIndex, version, lockTime, &index, &state)
							} else {
								log.Error("protocol error: unexpected TxRequest type")
								return
							}
						case messages.TxRequest_TXOUTPUT:
							if state == 1 { // Sending Outputs for InnerHash
								msg, err = sendOutputs(device, &outputs, &addressIndex, &coins, &hours, version, lockTime, &index, &state)
							} else {
								log.Error("protocol error: unexpected TxRequest type")
								return
							}
						case messages.TxRequest_TXFINISHED:
							if state == 3 {
								err = printSignatures(&msg)
								if err != nil {
									log.Error(err)
								}
								return
							}
							log.Error("protocol error: unexpected TXFINISHED message")
							return
						}
					case uint16(messages.MessageType_MessageType_Failure):
						failMsg, err := skyWallet.DecodeFailMsg(msg)
						if err != nil {
							log.Error(err)
							return
						}
						fmt.Printf("Failed with message: %s\n", failMsg)
						return
					case uint16(messages.MessageType_MessageType_ButtonRequest):
						msg, err = device.ButtonAck()
					default:
						log.Error("unexpected response message type from hardware wallet.")
						return
					}
				}
			} else {
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
					log.Error(err)
					return
				}

				for {
					switch msg.Kind {
					case uint16(messages.MessageType_MessageType_ResponseTransactionSign):
						signatures, err := skyWallet.DecodeResponseTransactionSign(msg)
						if err != nil {
							log.Error(err)
							return
						}
						fmt.Println(signatures)
						return
					case uint16(messages.MessageType_MessageType_Success):
						fmt.Println("Should end with ResponseTransactionSign request")
						return
					case uint16(messages.MessageType_MessageType_ButtonRequest):
						msg, err = device.ButtonAck()
						if err != nil {
							log.Error(err)
							return
						}
					case uint16(messages.MessageType_MessageType_PassphraseRequest):
						var passphrase string
						fmt.Printf("Input passphrase: ")
						fmt.Scanln(&passphrase)
						msg, err = device.PassphraseAck(passphrase)
						if err != nil {
							log.Error(err)
							return
						}
					case uint16(messages.MessageType_MessageType_PinMatrixRequest):
						var pinEnc string
						fmt.Printf("PinMatrixRequest response: ")
						fmt.Scanln(&pinEnc)
						msg, err = device.PinMatrixAck(pinEnc)
						if err != nil {
							log.Error(err)
							return
						}
					case uint16(messages.MessageType_MessageType_Failure):
						failMsg, err := skyWallet.DecodeFailMsg(msg)
						if err != nil {
							log.Error(err)
							return
						}

						fmt.Printf("Failed with message: %s\n", failMsg)
						return
					default:
						log.Errorf("received unexpected message type: %s", messages.MessageType(msg.Kind))
						return
					}
				}
			}
		},
	}
}

func sendInputs(device *skyWallet.Device, inputs *[]string, inputIndex *[]int, version int, lockTime int, index *int, state *int) (wire.Message, error) {
	var txInputs []*messages.TxAck_TransactionType_TxInputType
	startIndex := *index
	for i, input := range (*inputs)[*index:] {
		if len(txInputs) == 7 {
			return device.TxAck(txInputs, []*messages.TxAck_TransactionType_TxOutputType{}, version, lockTime)
		}
		var txInput messages.TxAck_TransactionType_TxInputType
		txInput.AddressN = []uint32{*proto.Uint32(uint32((*inputIndex)[startIndex+i]))}
		txInput.HashIn = proto.String(input)
		txInputs = append(txInputs, &txInput)
		*index++
	}
	if len(txInputs) != 0 {
		*index = 0
		*state++
		return device.TxAck(txInputs, nil, version, lockTime)
	} else if *index==len(*inputs){
		*index = 0
		*state++
	}
	return wire.Message{}, errors.New("empty inputs")
}

func sendOutputs(device *skyWallet.Device, outputs *[]string, addressIndex *[]int, coins *[]int64, hours *[]int64, version int, lockTime int, index *int, state *int) (wire.Message, error) {
	var txOutputs []*messages.TxAck_TransactionType_TxOutputType
	startIndex := *index
	for i, output := range (*outputs)[*index:] {
		if len(txOutputs) == 7 {
			return device.TxAck(nil, txOutputs, version, lockTime)
		}
		var txOutput messages.TxAck_TransactionType_TxOutputType
		txOutput.Address = proto.String(output)
		if i < len(*addressIndex) {
			txOutput.AddressN = []uint32{uint32((*addressIndex)[startIndex+i])}
		}
		txOutput.Coins = proto.Uint64(uint64((*coins)[startIndex+i]))
		txOutput.Hours = proto.Uint64(uint64((*hours)[startIndex+i]))
		txOutputs = append(txOutputs, &txOutput)
		*index++
	}
	if len(txOutputs) != 0 {
		*index = 0
		*state++
		return device.TxAck(nil, txOutputs, version, lockTime)
	} else if *index==len(*outputs){
		*index = 0
		*state++
	}
	return wire.Message{}, errors.New("empty outputs")
}

func printSignatures(msg *wire.Message) error {
	txRequest := &messages.TxRequest{}
	err := proto.Unmarshal(msg.Data, txRequest)
	if err != nil {
		return err
	}
	for _, sign := range txRequest.SignResult {
		fmt.Println(*sign.Signature)
	}
	return nil
}
