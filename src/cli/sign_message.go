package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	messages "github.com/skycoin/hardware-wallet-protob/go"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func signMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signMessage",
		Short: "Ask the device to sign a message using the secret key at given index.",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceType, _ := cmd.Flags().GetString("deviceType")
			addressN, _ := cmd.Flags().GetInt("addressN")
			message, _ := cmd.Flags().GetString("message")

			device := skyWallet.NewDevice(skyWallet.DeviceTypeFromString(deviceType))
			if device == nil {
				return fmt.Errorf("failed to create device")
			}
			defer device.Close()

			if os.Getenv("AUTO_PRESS_BUTTONS") == "1" && device.Driver.DeviceType() == skyWallet.DeviceTypeEmulator && runtime.GOOS == "linux" {
				err := device.SetAutoPressButton(true, skyWallet.ButtonRight)
				if err != nil {
					return err
				}
			}

			var signature string

			msg, err := device.SignMessage(addressN, message)
			if err != nil {
				return err
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				msg, err = device.ButtonAck()
				if err != nil {
					return err
				}
			}

			for msg.Kind != uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) && msg.Kind != uint16(messages.MessageType_MessageType_Failure) {
				if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
					var pinEnc string
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					msg, err = device.PinMatrixAck(pinEnc)
					if err != nil {
						return err
					}
					continue
				}

				if msg.Kind == uint16(messages.MessageType_MessageType_PassphraseRequest) {
					var passphrase string
					fmt.Printf("Input passphrase: ")
					fmt.Scanln(&passphrase)
					msg, err = device.PassphraseAck(passphrase)
					if err != nil {
						return err
					}
					continue
				}

				if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg, err = device.ButtonAck()
					if err != nil {
						return err
					}
					continue
				}
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) {
				signature, err = skyWallet.DecodeResponseSkycoinSignMessage(msg)
				if err != nil {
					return err
				}
			} else {
				failMsg, err := skyWallet.DecodeFailMsg(msg)
				if err != nil {
					return err
				}
				return fmt.Errorf("failed with: %s", failMsg)
			}

			fmt.Println(signature)
			return nil
		},
	}

	cmd.Flags().Int("addressN", 0, "Index of the address that will issue the signature. Assume 0 if not set.")
	cmd.Flags().String("message", "", "The message that the signature claims to be signing.")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
