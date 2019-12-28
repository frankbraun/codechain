package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/log"
)

// SigCtl implements the 'sigctl' command.
func SigCtl(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -m\n", argv0)
		fmt.Fprintf(os.Stderr, "Change signature control value.\n")
		fs.PrintDefaults()
	}
	m := fs.Int("m", -1, "Signature threshold M")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if *m == -1 {
		return fmt.Errorf("%s: option -m is mandatory", argv0)
	}
	if *m < 1 {
		return fmt.Errorf("%s: option -m must be >= 1", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	line, err := c.SignatureControl(*m)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}
