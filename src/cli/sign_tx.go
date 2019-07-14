package cli

import (
	// "fmt"
	"os"
	"runtime"

	"github.com/gogo/protobuf/proto"

	gcli "github.com/urfave/cli"

	messages "github.com/skycoin/hardware-wallet-protob/go"

	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func signTxCmd() gcli.Command {
	name := "signTx"
	return gcli.Command{
		Name:        name,
		Usage:       "Start a transaction with more than 8 inputs and 8 outputs.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.IntFlag{
				Name:  "outputs",
				Usage: "Number of outputs in transaction.",
			},
			gcli.IntFlag{
				Name:  "inputs",
				Usage: "Number of outputs in transaction.",
			},
			gcli.StringFlag{
				Name:  "coin",
				Usage: "The name of coin to use.",
			},
			gcli.IntFlag{
				Name:  "version",
				Usage: "Transaction version.",
			},
			gcli.IntFlag{
				Name:  "lockTime",
				Usage: "Transaction lock time.",
			},
			gcli.StringFlag{
				Name:  "txHash",
				Usage: "Transaction Hash",
			},
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			outputsCount := c.Int("outputs")
			inputsCount := c.Int("inputs")
			coinName := c.String("coin")
			version := c.Int("version")
			lockTime := c.Int("lockTime")
			txHash := []byte(c.String("txHash"))

			device := skyWallet.NewDevice(skyWallet.DeviceTypeFromString(c.String("deviceType")))
			log.Info("Start Action on signTx function")
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

			msg, err := device.SignTx(outputsCount, inputsCount, coinName, version, lockTime, txHash)
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
				log.Info(txRequest)
			default:
				log.Info("Unexpected response message")
			}
		},
	}
}
