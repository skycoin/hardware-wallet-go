/*
Package cli implements an interface for creating a CLI application.
Includes methods for manipulating wallets files and interacting with the
webrpc API to query a skycoin node's status.
*/
package cli

import (
	"fmt"

	gcli "github.com/urfave/cli"
)

const (
	// Version is the CLI Version
	Version = "0.24.1"
)

var (
	commandHelpTemplate = fmt.Sprintf(`USAGE:
        {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{if .Category}}

CATEGORY:
        {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
        {{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:
        {{range .VisibleFlags}}{{.}}
        {{end}}{{end}}
`)

	appHelpTemplate = fmt.Sprintf(`NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{.Copyright}}{{end}}
`)
)

// App Wraps the app so that main package won't use the raw App directly,
// which will cause import issue
type App struct {
	gcli.App
}

// NewApp creates an app instance
func NewApp() (*App, error) {
	gcli.AppHelpTemplate = appHelpTemplate
	gcli.SubcommandHelpTemplate = commandHelpTemplate
	gcli.CommandHelpTemplate = commandHelpTemplate

	gcliApp := gcli.NewApp()
	app := &App{
		App: *gcliApp,
	}

	commands := []gcli.Command{
		deviceSetMnemonicCmd(),
		deviceGenerateMnemonicCmd(),
		deviceAddressGenCmd(),
		deviceFirmwareUpdate(),
		deviceSignMessageCmd(),
		deviceCheckMessageSignatureCmd(),
		deviceSetPinCode(),
		deviceWipeCmd(),
		deviceBackupCmd(),
		deviceGetVersionCmd(),
		emulatorSetMnemonicCmd(),
		emulatorGenerateMnemonicCmd(),
		emulatorAddressGenCmd(),
		emulatorSignMessageCmd(),
		emulatorCheckMessageSignatureCmd(),
		emulatorSetPinCode(),
		emulatorWipeCmd(),
		emulatorBackupCmd(),
		emulatorGetVersionCmd(),
		sandbox(),
	}

	app.Name = "skycoin-cli"
	app.Version = Version
	app.Usage = "the skycoin command line interface"
	app.Commands = commands
	app.EnableBashCompletion = true
	app.OnUsageError = func(context *gcli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(context.App.Writer, "Error: %v\n\n", err)
		return gcli.ShowAppHelp(context)
	}
	app.CommandNotFound = func(ctx *gcli.Context, command string) {
		tmp := fmt.Sprintf("{{.HelpName}}: '%s' is not a {{.HelpName}} command. See '{{.HelpName}} --help'.\n", command)
		gcli.HelpPrinter(app.Writer, tmp, app)
		gcli.OsExiter(1)
	}

	return app, nil
}

// Run starts the app
func (app *App) Run(args []string) error {
	return app.App.Run(args)
}

func onCommandUsageError(command string) gcli.OnUsageErrorFunc {
	return func(c *gcli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(c.App.Writer, "Error: %v\n\n", err)
		return gcli.ShowCommandHelp(c, command)
	}
}
