package command

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/terminal"
)

func cleanSlate() error {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
outerA:
	for _, fi := range files {
		if def.ExcludePaths != nil {
			canonical := filepath.ToSlash(fi.Name())
			for _, excludePath := range def.ExcludePaths {
				if excludePath == canonical {
					continue outerA
				}
			}
		}
		if fi.IsDir() {
			fmt.Println(fi.Name() + "/")
		} else {
			fmt.Println(fi.Name())
		}
	}

	for {
		fmt.Print("delete all files and directories listed above? [y/n]: ")
		answer, err := terminal.ReadLine(os.Stdin)
		if err != nil {
			return err
		}
		a := string(bytes.ToLower(answer))
		if strings.HasPrefix(a, "y") {
			break
		} else if strings.HasPrefix(a, "n") {
			return errors.New("aborted")
		} else {
			fmt.Println("answer not recognized")
		}
	}

outerB:
	for _, fi := range files {
		if def.ExcludePaths != nil {
			canonical := filepath.ToSlash(fi.Name())
			for _, excludePath := range def.ExcludePaths {
				if excludePath == canonical {
					continue outerB
				}
			}
		}
		if err := os.RemoveAll(fi.Name()); err != nil {
			return err
		}
	}

	return nil
}

// CleanSlate implements the 'cleanslate' command.
func CleanSlate(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Remove all files except the .codechain dir and special files from current dir.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	return cleanSlate()
}
