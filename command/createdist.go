package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/log"
)

// CreateDist implements the 'createdist' command.
func CreateDist(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-f dist.tar.gz]\n", argv0)
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
	if *filename == "" {
		*filename = fmt.Sprintf("%x.tar.gz", c.Head())
	}
	if err := archive.CreateDist(c, *filename); err != nil {
		return err
	}
	fmt.Println(*filename)
	return nil
}
