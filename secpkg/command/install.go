package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/log"
)

// Install implements the secpkg 'install' command.
func Install(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s project.secpkg\n", argv0)
		fmt.Fprintf(os.Stderr, "Download, verify, and install package defined by project.secpkg.\n")
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
	// 1. Parse .secpkg file and validate it.
	pkg, err := secpkg.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	return pkg.Install()
}
