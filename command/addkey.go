package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/log"
)

// AddKey implements the 'addkey' command.
func AddKey(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-w] pubkey signature [comment]\n", argv0)
		fmt.Fprintf(os.Stderr, "Add new signer to hashchain.\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Be verbose")
	w := fs.Int("w", 1, "Signature weight w")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if *w < 1 {
		return fmt.Errorf("%s: option -w must be >= 1", argv0)
	}
	nArg := fs.NArg()
	if nArg != 2 && nArg != 3 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
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
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	var pubKey [32]byte
	copy(pubKey[:], pub)
	var signtr [64]byte
	copy(signtr[:], sig)
	line, err := c.AddKey(*w, pubKey, signtr, comment)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}
