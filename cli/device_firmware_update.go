package cli

import (
	"fmt"
	"io/ioutil"

	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/device-wallet"
)

func deviceFirmwareUpdate() gcli.Command {
	name := "deviceFirmwareUpdate"
	return gcli.Command{
		Name:        name,
		Usage:       "Update device's firmware.",
		Description: "",
		Flags: []gcli.Flag{
			gcli.StringFlag{
				Name:  "f, file",
				Usage: "path to the firmware .bin file",
			},
		},
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) {
			filePath := c.String("file")
			fmt.Printf("File : %s\n", filePath)
			firmware, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			deviceWallet.DeviceFirmwareUpload(firmware, make([]byte, 0))
		},
	}
}
