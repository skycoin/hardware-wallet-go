package cli

import (
	"errors"
)

func parseBool(s string) (*bool, error) {
	var usePassphrase bool
	switch passphrase {
	case "true":
		usePassphrase = true
	case "false":
		usePassphrase = false
	case "":
		return nil, nil
	default:
		return nil, errors.New("Invalid boolean argument")
	}
	return &usePassphrase, nil
}
