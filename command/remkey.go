package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/base64"
)

// RemKey implements the 'remkey' command.
func RemKey(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s pubkey\n", argv0)
		fmt.Fprintf(os.Stderr, "Remove existing signer from hashchain.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	pubkey := fs.Arg(0)
	pub, err := base64.Decode(pubkey, 32)
	if err != nil {
		return fmt.Errorf("cannot decode pubkey: %s", err)
	}
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	var pubKey [32]byte
	copy(pubKey[:], pub)
	line, err := c.RemoveKey(pubKey)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}
