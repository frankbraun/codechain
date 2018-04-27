package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/file"
)

// InitChain implements the 'init' command.
func InitChain() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	m := fs.Int("m", 1, "Signature threshold M")
	fs.Parse(os.Args[2:])
	if *m < 1 {
		return fmt.Errorf("%s: option -m must be >= 1", app)
	}
	if err := os.MkdirAll(codechainDir, 0700); err != nil {
		return err
	}
	exists, err := file.Exists(hashchainFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: file '%s' exists already", app, hashchainFile)
	}
	chain := hashchain.New(*m)
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
