package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/base64"
)

// AddKey implements the 'addkey' command.
func AddKey(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-w] pubkey signature [comment]\n", argv0)
		fmt.Fprintf(os.Stderr, "Add new signer to hashchain.\n")
		fs.PrintDefaults()
	}
	w := fs.Int("w", 1, "Signature weight W")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *w < 1 {
		return fmt.Errorf("%s: option -w must be >= 1", argv0)
	}
	nArg := fs.NArg()
	if nArg != 2 && nArg != 3 {
		fs.Usage()
		return flag.ErrHelp
	}
	pubkey := fs.Arg(0)
	pub, err := base64.Decode(pubkey, 32)
	if err != nil {
		return fmt.Errorf("cannot decode pubkey: %s", err)
	}
	signature := fs.Arg(1)
	sig, err := base64.Decode(signature, 64)
	if err != nil {
		return fmt.Errorf("cannot decode signature: %s", err)
	}
	var comment []byte
	if nArg == 3 {
		comment = []byte(fs.Arg(2))
	}
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	var pubKey [32]byte
	copy(pubKey[:], pub)
	var signtr [64]byte
	copy(signtr[:], sig)
	line, err := c.AddKey(pubKey, signtr, comment)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}
