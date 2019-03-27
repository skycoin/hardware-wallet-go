package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/src/device-wallet"
	messages "github.com/skycoin/hardware-wallet-go/src/device-wallet/messages/go"
)

func wipeCmd() gcli.Command {
	name := "wipe"
	return gcli.Command{
		Name:         name,
		Usage:        "Ask the device to wipe clean all the configuration it contains.",
		Description:  "",
		OnUsageError: onCommandUsageError(name),
		Flags: []gcli.Flag{
			gcli.StringFlag{
				Name:   "deviceType",
				Usage:  "Device type to send instructions to, hardware wallet (USB) or emulator.",
				EnvVar: "DEVICE_TYPE",
			},
		},
		Action: func(c *gcli.Context) {
			device := deviceWallet.NewDevice(deviceWallet.DeviceTypeFromString(c.String("deviceType")))
			if device == nil {
				return
			}

			msg, err := device.Wipe()
			if err != nil {
				log.Error(err)
				return
			}

			// get device connection instance
			dev, err := device.Driver.GetDevice()
			if err != nil {
				log.Error(err)
			}
			defer dev.Close()

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				msg, err = deviceWallet.DeviceButtonAck(dev)
				if err != nil {
					log.Error(err)
					return
				}
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				err = deviceWallet.Initialize(dev)
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
