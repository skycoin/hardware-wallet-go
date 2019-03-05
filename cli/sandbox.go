package cli

import (
	"fmt"

	"github.com/skycoin/hardware-wallet-go/device-wallet/wire"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
	"github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

func sandbox() gcli.Command {
	name := "sandbox"
	return gcli.Command{
		Name:         name,
		Usage:        "Sandbox.",
		Description:  "",
		Flags:        []gcli.Flag{},
		OnUsageError: onCommandUsageError(name),
		Action: func(_ *gcli.Context) {
			var deviceType deviceWallet.DeviceType
			if deviceWallet.DeviceConnected(deviceWallet.DeviceTypeEmulator) {
				deviceType = deviceWallet.DeviceTypeEmulator
			} else if deviceWallet.DeviceConnected(deviceType) {
				deviceType = deviceWallet.DeviceTypeUsb
			} else {
				log.Println("no device detected")
				return
			}

			_, err := deviceWallet.WipeDevice(deviceType)
			if err != nil {
				log.Error(err)
				return
			}

			_, err = deviceWallet.DeviceSetMnemonic(deviceType, "cloud flower upset remain green metal below cup stem infant art thank")
			if err != nil {
				log.Error(err)
				return
			}

			var pinEnc string
			var msg wire.Message
			msg, err = deviceWallet.DeviceChangePin(deviceType)
			if err != nil {
				log.Error(err)
				return
			}

			for msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				log.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				msg, err = deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
				if err != nil {
					log.Error(err)
					return
				}
			}

			// come on one-more time
			// testing what happen when we try to change an existing pin code
			msg, err = deviceWallet.DeviceChangePin(deviceType)
			if err != nil {
				log.Error(err)
				return
			}

			for msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				log.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				msg, err = deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
				if err != nil {
					log.Error(err)
					return
				}
			}

			msg, err = deviceWallet.DeviceAddressGen(deviceType, 9, 15, false)
			if err != nil {
				log.Error(err)
				return
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
				log.Printf("PinMatrixRequest response: ")
				fmt.Scanln(&pinEnc)
				msg, err = deviceWallet.DevicePinMatrixAck(deviceType, pinEnc)
				if err != nil {
					log.Error(err)
					return
				}

				if msg.Kind == uint16(messages.MessageType_MessageType_ResponseSkycoinAddress) {
					addresses, err := deviceWallet.DecodeResponseSkycoinAddress(msg)
					if err != nil {
						log.Error(err)
						return
					}
					log.Print("Successfully got address")
					log.Print(addresses)
				}
			} else {
				log.Println("Got addresses without pin code")
				addresses, err := deviceWallet.DecodeResponseSkycoinAddress(msg)
				if err != nil {
					log.Error(err)
					return
				}
				log.Print(addresses)
			}
		},
	}
}
