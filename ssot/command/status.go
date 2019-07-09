package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

func status() error {
	pkgDir := filepath.Join(homedir.SSOTPub(), "pkgs")
	pkgs, err := file.List(pkgDir)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		filename := filepath.Join(pkgDir, pkg, ssot.File)
		sh, err := ssot.Load(filename)
		if err != nil {
			return err
		}

		fmt.Printf("%s:\n", filename)
		fmt.Println(sh.MarshalText())
	}
	return nil
}

// Status implements the ssotpub 'status' command.
func Status(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Show status of managed packages.\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() > 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	if err := status(); err != nil {
		return err
	}
	return nil
}
