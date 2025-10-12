package cli

import (
	"fmt"
	"os"
	"runtime"

	messages "github.com/skycoin/hardware-wallet-protob/go"
	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func setMnemonicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setMnemonic",
		Short: "Configure the device with a mnemonic.",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceType, _ := cmd.Flags().GetString("deviceType")
			mnemonic, _ := cmd.Flags().GetString("mnemonic")

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

			msg, err := device.SetMnemonic(mnemonic)
			if err != nil {
				return err
			}

			if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
				msg, err = device.ButtonAck()
				if err != nil {
					return err
				}
			}

			responseMsg, err := skyWallet.DecodeSuccessOrFailMsg(msg)
			if err != nil {
				return err
			}

			fmt.Println(responseMsg)
			return nil
		},
	}

	cmd.Flags().String("mnemonic", "", "Mnemonic that will be stored in the device to generate addresses.")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
