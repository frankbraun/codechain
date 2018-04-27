package command

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"golang.org/x/crypto/ed25519"
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
	pub, err := base64.URLEncoding.DecodeString(pubkey)
	if err != nil {
		return fmt.Errorf("cannot decode pubkey: %s", err)
	}
	signature := fs.Arg(1)
	sig, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("cannot decode signature: %s", err)
	}
	var comment string
	if nArg == 3 {
		comment = fs.Arg(2)
	}
	if !ed25519.Verify(pub, append(pub, []byte(comment)...), sig) {
		return fmt.Errorf("signature does not verify")
	}
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	if err := c.Verify(); err != nil {
		return err
	}
	line, err := c.AddKey(hashchainFile, pubkey, signature, comment)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}
