package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/file"
)

// Start implements the 'start' command.
func Start(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-m]\n", argv0)
		fmt.Fprintf(os.Stderr, "Initialized new .codechain/hashchain in current directory.\n")
		fs.PrintDefaults()
	}
	m := fs.Int("m", 1, "Signature threshold M")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *m < 1 {
		return fmt.Errorf("%s: option -m must be >= 1", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := os.MkdirAll(codechainDir, 0700); err != nil {
		return err
	}
	exists, err := file.Exists(hashchainFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: file '%s' exists already", argv0, hashchainFile)
	}
	chain, err := hashchain.New(*m)
	if err != nil {
		return err
	}
	if err := chain.Save(os.Stdout); err != nil {
		return err
	}
	f, err := os.Create(hashchainFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return chain.Save(f)
}
