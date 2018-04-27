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
func AddKey() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ContinueOnError)
	w := fs.Int("w", 1, "Signature weight W")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if *w < 1 {
		return fmt.Errorf("%s: option -w must be >= 1", app)
	}
	nArg := fs.NArg()
	if nArg != 2 && nArg != 3 {
		return fmt.Errorf("%s: expecting args: pubkey signature [comment]", app)
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
	return c.AddKey(hashchainFile, pubkey, signature, comment)
}
