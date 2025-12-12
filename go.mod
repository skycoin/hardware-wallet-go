module github.com/skycoin/hardware-wallet-go

go 1.25.1

require (
	github.com/gogo/protobuf v1.3.2
	github.com/skycoin/hardware-wallet-protob v0.0.0-20250805154629-410561e1bc2f
	github.com/skycoin/skycoin v0.28.1-0.20251012182647-a1a88ea0df8f //DO NOT MODIFY OR UPDATE v0.28.1-0.20251012182647-a1a88ea0df8f
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/google/gousb v1.1.3
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// IT IS FORBIDDEN TO USE REPLACE DIRECTIVES

// [error] The go.mod file for the module providing named packages contains one or
//	more replace directives. It must not contain directives that would cause
//	it to be interpreted differently than if it were the main module.

// Uncomment for tests with local sources
//replace github.com/skycoin/hardware-wallet-protob => ../hardware-wallet-protob
//replace github.com/skycoin/skycoin => ../skycoin

// Below should reflect current versions of the following deps
// To update deps to specific commit hash:
// 1) Uncomment one of the following lines and substituite version with desired commit hash:
//replace github.com/skycoin/skycoin => github.com/skycoin/skycoin v0.28.1-0.20251205225511-c088af7bbed1
//replace github.com/skycoin/hardware-wallet-protob => github.com/skycoin/hardware-wallet-protob v0.0.0-20250805154629-410561e1bc2f
// 2) Run `go mod tidy && go mod vendor`
// 3) Copy the populated version string to the correct place in require(...) above - replacing the specified version string
// 4) Re-comment the uncommented replace directive above
// 5) Save this file.
// 6) Run `go mod tidy && go mod vendor`
