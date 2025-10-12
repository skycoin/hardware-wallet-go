package cli

import (

	"github.com/spf13/cobra"
	skyWallet "github.com/skycoin/hardware-wallet-go/src/skywallet"
)

func getUsbDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getUsbDetails",
		Short: "Ask host usb about details for the hardware wallet",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceType, _ := cmd.Flags().GetString("deviceType")
			device := skyWallet.NewDevice(skyWallet.DeviceTypeFromString(deviceType))
			if device == nil {
				return nil
			}
			defer device.Close()

			infos, err := device.GetUsbInfo()
			if err != nil {
				log.Errorln(err)
			}
			for infoIdx := range infos {
				log.Infoln("-----------------------------------------")
				if infos[infoIdx].VendorID == skyWallet.SkycoinVendorID {
					log.Printf("%-13d%-5s%s", infos[infoIdx].VendorID, "==>", "Skycoin Foundation")
				}
				if infos[infoIdx].ProductID == skyWallet.SkycoinHwProductID {
					log.Printf("%-13d%-5s%s", infos[infoIdx].ProductID, "==>", "Hardware Wallet")
				}
				log.Printf("%-13s%-5s%s", "Device path", "==>", infos[infoIdx].Path)
			}
			return nil
		},
	}

	cmd.Flags().String("deviceType", "", "Device type to send instructions to, hardware wallet (USB) or emulator.")
	return cmd
}
