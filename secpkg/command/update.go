package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

// updateAll updates all packages. If an update for a single fails the error
// is reported on stderr and the next package will be updated.
// The function returns the first encountered error, if any.
func updateAll(ctx context.Context) error {
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs")
	pkgs, err := file.List(pkgDir)
	if err != nil {
		return err
	}
	var firstError error
	for _, pkg := range pkgs {
		fmt.Printf("updating package '%s'\n", pkg)
		if err := secpkg.Update(ctx, secpkg.NewResolver(), pkg); err != nil {
			if firstError == nil {
				firstError = err
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
	return firstError
}

// Update implements the secpkg 'update' command.
func Update(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-all] [project_name]\n", argv0)
		fmt.Fprintf(os.Stderr, "Update installed package with given project_name, if necessary.\n")
		fs.PrintDefaults()
	}
	all := fs.Bool("all", false, "Update all installed packages")
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
	// do not check for codechain update during codechain update
	if *all || fs.Arg(0) != "codechain" {
		if err := secpkg.UpToDate("codechain"); err != nil {
			return err
		}
	}
	if *all {
		return updateAll(context.Background())
	}
	return secpkg.Update(context.Background(), secpkg.NewResolver(), fs.Arg(0))
}
