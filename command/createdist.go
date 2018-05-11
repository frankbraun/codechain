package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
)

func createDist(c *hashchain.HashChain, filename string) error {
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("distribution file '%s' exists already", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("creating distribution '%s'", filename)
	return archive.Create(f, c, def.PatchDir)
}

// CreateDist implements the 'createdist' command.
func CreateDist(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -f dist.tar.gz\n", argv0)
		fmt.Fprintf(os.Stderr, "Create distribution file (for `codechain apply -f`).\n")
		fs.PrintDefaults()
	}
	filename := fs.String("f", "", "Distribution file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if *filename == "" {
		return fmt.Errorf("%s: option -f is mandatory", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	return createDist(c, *filename)
}
