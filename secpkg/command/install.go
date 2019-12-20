package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

func install(pkgFlag bool, name string) error {
	ctx := context.Background()
	if pkgFlag {
		// make sure codechain is actually installed
		if _, err := secpkg.CheckUpdate(ctx, "codechain"); err != nil {
			if err == secpkg.ErrNotInstalled {
				fmt.Fprintf(os.Stderr, "you must install codechain via `secpkg install` in order to use option -p\n")
			}
			return err
		}
		securePackageDir := filepath.Join(homedir.SecPkg(), "pkgs", "codechain",
			"src", "packages")
		name = filepath.Join(securePackageDir, name+".secpkg")
	}
	// 1. Parse .secpkg file and validate it.
	pkg, err := secpkg.Load(name)
	if err != nil {
		return err
	}
	return pkg.Install(ctx)
}

// Install implements the secpkg 'install' command.
func Install(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-p] project.secpkg\n", argv0)
		fmt.Fprintf(os.Stderr, "Download, verify, and install package defined by project.secpkg.\n")
		fs.PrintDefaults()
	}
	pkgFlag := fs.Bool("p", false, "Install secure package file of given name distributed by Codechain")
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
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	return install(*pkgFlag, fs.Arg(0))
}
