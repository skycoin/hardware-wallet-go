package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/skycoin/hardware-wallet-go/src/skywallet/wire"

	"github.com/gogo/protobuf/proto"

	gcli "github.com/urfave/cli"

	messages "github.com/skycoin/hardware-wallet-protob/go"

	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
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
			case skyWallet.BitcoinCoinType:
				err = transactionBitcoinSign(device, prevHash, outputs, coins, inputIndex, addressIndex)
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

	coinName := "Skycoin"
	version := 1
	lockTime := 0
	txHash := "dkdji9e2oidhash"

	if len(inputs) != len(inputIndex) {
		return fmt.Errorf("Every given input hash should have the an inputIndex")
	}
	if len(outputs) != len(hours) {
		return fmt.Errorf("Every given output should have a coin value")
	}

	if len(inputs) > 7 || len(outputs) > 7 {
		state := 0
		index := 0

		msg, err := device.SignTx(len(outputs), len(inputs), coinName, version, lockTime, txHash)

		for {
			if err != nil {
				return err
			}
			switch msg.Kind {
			case uint16(messages.MessageType_MessageType_TxRequest):
				txRequest := &messages.TxRequest{}
				err = proto.Unmarshal(msg.Data, txRequest)
				if err != nil {
					return err
				}
				switch *txRequest.RequestType {
				case messages.TxRequest_TXINPUT:
					if state == 0 { // Sending Inputs for InnerHash
						msg, err = sendInputs(device, &inputs, &inputIndex, version, lockTime, &index, &state)
					} else if state == 2 { // Sending Inputs for Signatures
						err = printSignatures(&msg)
						if err != nil {
							return err
						}
						msg, err = sendInputs(device, &inputs, &inputIndex, version, lockTime, &index, &state)
					} else {
						return fmt.Errorf("protocol error: unexpected TxRequest type")
					}
				case messages.TxRequest_TXOUTPUT:
					if state == 1 { // Sending Outputs for InnerHash
						msg, err = sendOutputs(device, &outputs, &addressIndex, &coins, &hours, version, lockTime, &index, &state)
					} else {
						return fmt.Errorf("protocol error: unexpected TxRequest type")
					}
				case messages.TxRequest_TXFINISHED:
					if state == 3 {
						err = printSignatures(&msg)
						return err
					}
					return fmt.Errorf("protocol error: unexpected TXFINISHED message")
				}
			case uint16(messages.MessageType_MessageType_Failure):
				failMsg, err := skyWallet.DecodeFailMsg(msg)
				if err != nil {
					return err
				}
				return fmt.Errorf("Failed with message: %s", failMsg)
			case uint16(messages.MessageType_MessageType_ButtonRequest):
				msg, err = device.ButtonAck()
			default:
				return fmt.Errorf("unexpected response message type from hardware wallet")
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
				return fmt.Errorf("Should end with ResponseTransactionSign request")
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
					return err
				}

				return fmt.Errorf("Failed with message: %s", failMsg)
			default:
				return fmt.Errorf("received unexpected message type: %s", messages.MessageType(msg.Kind))
			}
		}
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
	} else if *index == len(*inputs) {
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
	} else if *index == len(*outputs) {
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

func transactionBitcoinSign(device *skyWallet.Device, prevHashes, outputs []string, coins []int64, inputIndex, addressIndex []int) error {

	coinName := "Bitcoin"
	version := 1
	lockTime := 0
	txHash := "dkdji9e2oidhash"

	if len(prevHashes) != len(inputIndex) {
		return fmt.Errorf("Every given input index should have a hash of previous the tx")
	}

	state := 0
	index := 0

	var transactionInputs []*messages.BitcoinTransactionInput
	var transactionOutputs []*messages.BitcoinTransactionOutput
	for i, prevHash := range prevHashes {
		var transactionInput messages.BitcoinTransactionInput
		transactionInput.AddressN = proto.Uint32(uint32(inputIndex[i]))
		decoded, err := hex.DecodeString(prevHash)
		if err != nil {
			return err
		}
		transactionInput.PrevHash = decoded
		transactionInputs = append(transactionInputs, &transactionInput)
	}
	for i, output := range outputs {
		var transactionOutput messages.BitcoinTransactionOutput
		transactionOutput.Address = proto.String(output)
		transactionOutput.Coin = proto.Uint64(uint64(coins[i]))
		if i < len(addressIndex) {
			transactionOutput.AddressIndex = proto.Uint32(uint32(addressIndex[i]))
		}
		transactionOutputs = append(transactionOutputs, &transactionOutput)
	}

	msg, err := device.SignTx(len(outputs), len(prevHashes), coinName, version, lockTime, txHash)

	for {
		if err != nil {
			return err
		}
		switch msg.Kind {
		case uint16(messages.MessageType_MessageType_TxRequest):
			txRequest := &messages.TxRequest{}
			err = proto.Unmarshal(msg.Data, txRequest)
			if err != nil {
				return err
			}
			switch *txRequest.RequestType {
			case messages.TxRequest_TXOUTPUT:
				if state == 0 { // Sending Outputs for Confirmation
					msg, err = sendBitcoinOutputs(device, transactionOutputs, &index, &state)
				} else {
					return fmt.Errorf("protocol error: unexpected TxRequest type")
				}
			case messages.TxRequest_TXINPUT:
				if state == 1 {
					err = printSignatures(&msg)
					if err != nil {
						return err
					}
					msg, err = sendBitcoinInputs(device, transactionInputs, &index, &state)
				} else {
					return fmt.Errorf("protocol error: unexpected TxRequest type")
				}
			case messages.TxRequest_TXFINISHED:
				if state == 2 {
					err = printSignatures(&msg)
					return err
				}
				return fmt.Errorf("protocol error: unexpected TXFINISHED message")
			}
		case uint16(messages.MessageType_MessageType_Failure):
			failMsg, err := skyWallet.DecodeFailMsg(msg)
			if err != nil {
				return err
			}
			return fmt.Errorf("Failed with message: %s", failMsg)
		case uint16(messages.MessageType_MessageType_ButtonRequest):
			msg, err = device.ButtonAck()
		default:
			return fmt.Errorf("unexpected response message type from hardware wallet")
		}
	}
}

func sendBitcoinOutputs(device *skyWallet.Device, outputs []*messages.BitcoinTransactionOutput, index, state *int) (wire.Message, error) {
	var txOutputs []*messages.BitcoinTransactionOutput
	for i, output := range outputs[*index:] {
		if i == 7 {
			return device.BitcoinTxAck(nil, txOutputs)
		}
		txOutputs = append(txOutputs, output)
		*index++
	}
	if len(txOutputs) != 0 {
		*index = 0
		*state++
		return device.BitcoinTxAck(nil, txOutputs)
	} else if *index == len(outputs) {
		*index = 0
		*state++
	}
	return wire.Message{}, errors.New("empty outputs")
}

func sendBitcoinInputs(device *skyWallet.Device, inputs []*messages.BitcoinTransactionInput, index, state *int) (wire.Message, error) {
	var txInputs []*messages.BitcoinTransactionInput
	for i, input := range inputs[*index:] {
		if i == 7 {
			return device.BitcoinTxAck(txInputs, nil)
		}
		txInputs = append(txInputs, input)
		*index++
	}
	if len(txInputs) != 0 {
		*index = 0
		*state++
		return device.BitcoinTxAck(txInputs, nil)
	} else if *index == len(inputs) {
		*index = 0
		*state++
	}
	return wire.Message{}, errors.New("empty inputs")
}
