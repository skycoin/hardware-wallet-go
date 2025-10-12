package cli

import (
	"fmt"
	"os"
	"runtime"

	messages "github.com/skycoin/hardware-wallet-protob/go"
	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func generateMnemonicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generateMnemonic",
		Short: "Ask the device to generate a mnemonic and configure itself with it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			usePassphrase, _ := cmd.Flags().GetBool("usePassphrase")
			wordCount, _ := cmd.Flags().GetInt("wordCount")
			deviceType, _ := cmd.Flags().GetString("deviceType")

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

			msg, err := device.GenerateMnemonic(uint32(wordCount), usePassphrase)
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

	cmd.Flags().Bool("usePassphrase", false, "Configure a passphrase")
	cmd.Flags().Int("wordCount", 12, "Use a specific (12 | 24) number of words for the Mnemonic")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
