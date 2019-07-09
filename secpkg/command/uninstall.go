package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/log"
)

// Uninstall implements the secpkg 'uninstall' command.
func Uninstall(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s project_name\n", argv0)
		fmt.Fprintf(os.Stderr, "Uninstall package with given project_name.\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	if fs.Arg(0) != "codechain" {
		if err := secpkg.UpToDate("codechain"); err != nil {
			return err
		}
	}
	return secpkg.Uninstall(fs.Arg(0))
}
