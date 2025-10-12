package cli

import (
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"
)

const (
	Version = "1.7.0"
)

var log = logging.MustGetLogger("skycoin-hw-cli")

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "skycoin-hw-cli",
		Short:   "the skycoin hardware wallet command line interface",
		Version: Version,
	}

	rootCmd.AddCommand(
		applySettingsCmd(),
		setMnemonicCmd(),
		featuresCmd(),
		generateMnemonicCmd(),
		addressGenCmd(),
		firmwareUpdate(),
		signMessageCmd(),
		checkMessageSignatureCmd(),
		setPinCode(),
		removePinCode(),
		wipeCmd(),
		backupCmd(),
		recoveryCmd(),
		cancelCmd(),
		transactionSignCmd(),
		getRawEntropyCmd(),
		getMixedEntropyCmd(),
		getUsbDetails(),
	)

	return rootCmd
}
