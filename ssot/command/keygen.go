package command

import (
	"github.com/frankbraun/codechain/command"
	"github.com/frankbraun/codechain/util/homedir"
)

// KeyGen implements the ssotpub 'keygen' command.
func KeyGen(argv0 string, args ...string) error {
	return command.KeyGen(homedir.SSOTPub(), argv0, args...)
}
