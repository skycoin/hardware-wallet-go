package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func firmwareUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firmwareUpdate",
		Short: "Update device's firmware.",
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

			err := device.FirmwareUpload(nil, [32]byte{})
			if err != nil {
				return err
			}

			fmt.Println("Firmware uploaded successfully")
			return nil
		},
	}

	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")
	return cmd
}
