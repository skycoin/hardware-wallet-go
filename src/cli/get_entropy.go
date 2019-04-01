package cli

import (
	gcli "github.com/urfave/cli"

	deviceWallet "github.com/skycoin/hardware-wallet-go/src/device-wallet"
)

func getEntropyCmd() gcli.Command {
	name := "getEntropy"
	return gcli.Command{
		Name:         name,
		ShortName:    "",
		Aliases:      nil,
		Usage:        "Get internal entropy from the device and write it down to a file",
		UsageText:    "",
		Description:  "",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action: func(c *gcli.Context) {
			entropyBytes := uint32(c.Int("entropyBytes"))
			outFile := c.String("outFile")
			if len(outFile) == 0 {
				log.Error("outFile is mandatory")
				return
			}
			device := deviceWallet.NewDevice(deviceWallet.DeviceTypeUSB)
			if device == nil {
				return
			}
			dev, err := device.Driver.GetDevice()
			if err != nil {
				log.Error("unable to open the device", err)
				return
			}
			defer dev.Close()
			log.Infoln("Getting entropy from device", outFile)
			err = device.SaveDeviceEntropyInFile(dev, outFile, entropyBytes)
			if err != nil {
				log.Error(err)
				return
			}
		},
		OnUsageError: onCommandUsageError(name),
		Subcommands:  nil,
		Flags: []gcli.Flag{
			gcli.IntFlag{
				Name:  "entropyBytes",
				Value: 1048576,
				Usage: "Total number of how many bytes of entropy to read.",
			},
			gcli.StringFlag{
				Name:  "outFile",
				Usage: "File path to write out the entropy buffers.",
			},
		},
		SkipFlagParsing:    false,
		SkipArgReorder:     false,
		HideHelp:           false,
		Hidden:             false,
		HelpName:           "",
		CustomHelpTemplate: "",
	}
}
