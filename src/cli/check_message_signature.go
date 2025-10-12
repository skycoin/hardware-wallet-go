package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func checkMessageSignatureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkMessageSignature",
		Short: "Check a message signature matches the given address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceType, _ := cmd.Flags().GetString("deviceType")
			message, _ := cmd.Flags().GetString("message")
			signature, _ := cmd.Flags().GetString("signature")
			address, _ := cmd.Flags().GetString("address")

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

			msg, err := device.CheckMessageSignature(message, signature, address)
			if err != nil {
				return err
			}

			responseMsg, err := skyWallet.DecodeSuccessOrFailMsg(msg)
			if err != nil {
				return err
			}

			fmt.Println(responseMsg)
			return nil
		},
	}

	cmd.Flags().String("message", "", "The message that the signature claims to be signing.")
	cmd.Flags().String("signature", "", "Signature of the message.")
	cmd.Flags().String("address", "", "Address to verify against the signature.")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
