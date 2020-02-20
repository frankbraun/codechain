package command

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/gnumake"
	"github.com/frankbraun/codechain/util/log"
)

func containsFile(dir string) (bool, error) {
	var containsFile bool
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if containsFile {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			containsFile = true
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return containsFile, nil
}

func testBuild() error {
	log.Println("test build")

	// 1. Create temporary directory TMPDIR with `build` and `local` subdirectories.
	log.Println("1. Create temporary directory TMPDIR with `build` and `local` subdirectories.")
	dir, err := ioutil.TempDir("", "testbuild")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	buildDir := filepath.Join(dir, "build")
	if os.Mkdir(buildDir, 0755); err != nil {
		return err
	}
	localDir := filepath.Join(dir, "local")
	if os.Mkdir(localDir, 0755); err != nil {
		return err
	}

	// 2. `mkdir TMPDIR/build/.codechain`
	log.Println("2. `mkdir TMPDIR/build/.codechain`")
	codechainDir := filepath.Join(buildDir, def.DefaultCodechainDir)
	if err := os.Mkdir(codechainDir, 0755); err != nil {
		return err
	}

	// 3. `cp .codechain/hashchain TMPDIR/build/.codechain`
	log.Println("3. `cp .codechain/hashchain TMPDIR/build/.codechain`")
	err = file.Copy(def.HashchainFile, filepath.Join(codechainDir, "hashchain"))
	if err != nil {
		return err
	}

	// 4. `cp -r .codechain/patches TMPDIR/build/.codechain`
	log.Println("4. `cp -r .codechain/patches TMPDIR/build/.codechain`")
	err = file.CopyDir(def.PatchDir, filepath.Join(codechainDir, "patches"))
	if err != nil {
		return err
	}

	// 5. `cd TMPDIR/build`
	log.Println("5. `cd TMPDIR/build`")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(buildDir); err != nil {
		return err
	}
	defer os.Chdir(cwd)

	// 6. `codechain apply`
	log.Println("6. `codechain apply`")
	c, err := hashchain.ReadFile(def.UnoverwriteableHashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.Apply(nil, def.UnoverwriteablePatchDir); err != nil {
		return err
	}

	// 7. `make prefix=TMPDIR/local`
	log.Println("7. `make prefix=TMPDIR/local`")
	if err := gnumake.Call(localDir); err != nil {
		return err
	}

	// 8. `make prefix=TMPDIR/local install`
	log.Println("8. `make prefix=TMPDIR/local install`")
	if err := gnumake.Install(localDir); err != nil {
		return err
	}

	// 9. Make sure TMPDIR/local contains at least one file.
	log.Println("9. Make sure TMPDIR/local contains at least one file.")
	contains, err := containsFile(localDir)
	if err != nil {
		return err
	}
	if !contains {
		return errors.New("'make install' doesn't install any files")
	}

	// 10. `make prefix=TMPDIR/local uninstall`
	log.Println("10. `make prefix=TMPDIR/local uninstall`")
	if err := gnumake.Uninstall(localDir); err != nil {
		return err
	}

	// 11. Make sure TMPDIR/local contains no files (but empty directories are OK).
	log.Println("11. Make sure TMPDIR/local contains no files (but empty directories are OK).")
	contains, err = containsFile(localDir)
	if err != nil {
		return err
	}
	if contains {
		return errors.New("'make uninstall' leaves files")
	}

	// 12. Delete temporary directory TMPDIR.
	log.Println("12. Delete temporary directory TMPDIR.")
	return nil
}

// TestBuild implements the ssotpub 'testbuild' command.
func TestBuild(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Test package build.\n")
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
	if err := testBuild(); err != nil {
		return err
	}
	return nil
}
