package command

import (
	"github.com/frankbraun/codechain/command"
	"github.com/frankbraun/codechain/util/homedir"
)

// KeyFile implements the ssotpub 'keyfile' command.
func KeyFile(argv0 string, args ...string) error {
	return command.KeyFile("codechain", homedir.SSOTPub(), argv0, args...)
}
