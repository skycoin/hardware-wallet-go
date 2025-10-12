package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func getRawEntropyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getRawEntropy",
		Short: "Get device raw internal entropy and write it down to a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceType, _ := cmd.Flags().GetString("deviceType")
			entropyBytes, _ := cmd.Flags().GetInt("entropyBytes")

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

			entropy, err := skyWallet.MessageDeviceGetRawEntropy(uint32(entropyBytes))
			if err != nil {
				return err
			}

			var entropyData []byte
			for _, chunk := range entropy {
				entropyData = append(entropyData, chunk[:]...)
			}

			err = ioutil.WriteFile("/tmp/entropy.dump", entropyData, 0644)
			if err != nil {
				return err
			}

			fmt.Println("Raw entropy dumped to: /tmp/entropy.dump")
			return nil
		},
	}

	cmd.Flags().Int("entropyBytes", 1048576, "Number of how many bytes of entropy to read.")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")

	return cmd
}
