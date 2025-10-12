package cli

import (
	"fmt"
	"os"
	"runtime"

	messages "github.com/skycoin/hardware-wallet-protob/go"

	"github.com/spf13/cobra"

	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func wipeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wipe",
		Short: "Ask the device to wipe clean all the configuration it contains.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			msg, err := device.Wipe()
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

			if len(responseMsg) > 0 {
				fmt.Println(responseMsg)
			} else {
				fmt.Println("Firmware was successfully wiped from the device")
			}
			return nil
		},
	}

	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
