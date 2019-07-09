package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/terminal"
)

func cleanSlate() error {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
outer:
	for _, fi := range files {
		if def.ExcludePaths != nil {
			canonical := filepath.ToSlash(fi.Name())
			for _, excludePath := range def.ExcludePaths {
				if excludePath == canonical {
					continue outer
				}
			}
		}
		if fi.IsDir() {
			fmt.Println(fi.Name() + "/")
		} else {
			fmt.Println(fi.Name())
		}
	}

	err = terminal.Confirm("delete all files and directories listed above?")
	if err != nil {
		return err
	}

	return file.RemoveAll(".", def.ExcludePaths)
}

// CleanSlate implements the 'cleanslate' command.
func CleanSlate(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Remove all files except the .codechain dir and special files from current dir.\n")
		fs.PrintDefaults()
	}
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
	return cleanSlate()
}
