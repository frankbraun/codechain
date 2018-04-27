package command

import "github.com/frankbraun/codechain/hashchain"

// Verify implement the 'verify' command.
func Verify() error {
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	return c.Verify()
}
