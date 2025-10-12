package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	messages "github.com/skycoin/hardware-wallet-protob/go"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func applySettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "applySettings",
		Short: "Apply settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			passphrase, _ := cmd.Flags().GetString("usePassphrase")
			label, _ := cmd.Flags().GetString("label")
			language, _ := cmd.Flags().GetString("language")
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

			usePassphrase, _err := parseBool(passphrase)
			if _err != nil {
				return fmt.Errorf("valid values for usePassphrase are true or false")
			}
			msg, err := device.ApplySettings(usePassphrase, label, language)
			if err != nil {
				return err
			}

			for msg.Kind != uint16(messages.MessageType_MessageType_Failure) && msg.Kind != uint16(messages.MessageType_MessageType_Success) {
				if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
					msg, err = device.ButtonAck()
					if err != nil {
						return err
					}
					continue
				}

				if msg.Kind == uint16(messages.MessageType_MessageType_PinMatrixRequest) {
					var pinEnc string
					fmt.Printf("PinMatrixRequest response: ")
					fmt.Scanln(&pinEnc)
					pinAckResponse, err := device.PinMatrixAck(pinEnc)
					if err != nil {
						return err
					}
					log.Infof("PinMatrixAck response: %s", pinAckResponse)
					continue
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

	cmd.Flags().String("usePassphrase", "", "Configure a passphrase (true or false)")
	cmd.Flags().String("label", "", "Configure a device label")
	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")
	cmd.Flags().String("language", "", "Configure a device language")

	return cmd
}
