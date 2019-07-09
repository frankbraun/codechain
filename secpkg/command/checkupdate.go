package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

// checkUpdateAll checks all packages for updates. If an update check for a
// single fails the error is reported on stderr and the next package will be
// updated. The function returns the first encountered error, if any.
func checkUpdateAll(ctx context.Context) error {
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs")
	pkgs, err := file.List(pkgDir)
	if err != nil {
		return err
	}
	var firstError error
	for _, pkg := range pkgs {
		fmt.Printf("%s: checking\n", pkg)
		needsUpdate, err := secpkg.CheckUpdate(ctx, pkg)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		} else if needsUpdate {
			fmt.Println(color.RedString("needs update!"))
		} else {
			fmt.Println("up-to-date")
		}
	}
	return firstError
}

// CheckUpdate implements the secpkg 'checkupdate' command.
func CheckUpdate(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-all] [project_name]\n", argv0)
		fmt.Fprintf(os.Stderr, "Check if installed package with given project_name can be updated.\n")
		fs.PrintDefaults()
	}
	all := fs.Bool("all", false, "Check all installed packages")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if *all && fs.NArg() != 0 || !*all && fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	if *all {
		return checkUpdateAll(context.Background())
	}
	needsUpdate, err := secpkg.CheckUpdate(context.Background(), fs.Arg(0))
	if err != nil {
		return err
	}
	if needsUpdate {
		fmt.Println(color.RedString("needs update!"))
	} else {
		fmt.Println("up-to-date")
	}
	return nil
}
