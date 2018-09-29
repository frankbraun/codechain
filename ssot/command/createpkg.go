package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func createPkg(c *hashchain.HashChain, name, dns, url, secKeyFile string) error {
	head := c.Head()
	fmt.Printf("create package for head %x\n", head)
	_, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}
	// Create package
	pkg, err := secpkg.New(name, dns, url, head)
	if err != nil {
		return err
	}
	// Make sure the project has not been published before
	pkgDir := filepath.Join(homedir.SSOTPub(), "pkg", pkg.Name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("package has already been published: '%s' exists", pkgDir)
	}
	// Create .secpkg file
	exists, err = file.Exists(secpkg.File)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("secure package already exists: %s", secpkg.File)
	}
	err = ioutil.WriteFile(secpkg.File, []byte(pkg.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("'%s' file has been written\n", secpkg.File)

	return nil
}

// CreatePKG implements the ssotpub 'createpkg' command.
func CreatePkg(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -name name -dns FQDN -url URL -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Create secure package and first signed head.\n")
		fs.PrintDefaults()
	}
	name := fs.String("name", "", "The project's package name")
	dns := fs.String("dns", "", "Fully qualified comain name for _codechain TXT records (SSOT)")
	url := fs.String("url", "", "URL to download project files from (URL/head.tar.gz)")
	secKey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("%s: option -name is mandatory", argv0)
	}
	if *dns == "" {
		return fmt.Errorf("%s: option -dns is mandatory", argv0)
	}
	if *url == "" {
		return fmt.Errorf("%s: option -url is mandatory", argv0)
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if err := seckey.Check(homedir.SSOTPub(), *secKey); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	// add interrupt handler
	interrupt.AddInterruptHandler(func() {
		c.Close()
	})
	// run createPkg
	go func() {
		if err := createPkg(c, *name, *dns, *url, *secKey); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
