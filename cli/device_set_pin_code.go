package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func deviceSetPinCode() gcli.Command {
	name := "deviceSetPinCode"
	return gcli.Command{
		Name:         name,
		Usage:        "Configure a PIN code on a device.",
		Description:  "",
		Flags:        []gcli.Flag{},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			var pinEnc string
			kind, _ := deviceWallet.DeviceChangePin(deviceWallet.DeviceTypeUsb)
			for kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				fmt.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				kind, _ = deviceWallet.DevicePinMatrixAck(deviceWallet.DeviceTypeUsb, pinEnc)
			}
		},
	}
}
