package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func createPkg(c *hashchain.HashChain, name, dns, URL, secKeyFile string) error {
	head := c.Head()
	fmt.Printf("create package for head %x\n", head)
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}
	// 2. Create package (before 1., because this checks the arguments)
	if _, err := url.Parse(URL); err != nil {
		return err
	}
	pkg, err := secpkg.New(name, dns, head)
	if err != nil {
		return err
	}

	// 1. Make sure the project has not been published before
	pkgDir := filepath.Join(homedir.SSOTPub(), "pkgs", pkg.Name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("package already published: '%s' exists", pkgDir)
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
	fmt.Printf("%s: written\n", secpkg.File)

	// 3. Create the first signed head with counter set to 0.
	sh := ssot.SignHead(head, 0, *secKey)

	// 4. Create the directory ~/.config/ssotpub/pkgs/NAME/dists
	//    and save the current distribution to
	//    ~/.config/ssotpub/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`)
	distDir := filepath.Join(pkgDir, "dists")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	distFile := filepath.Join(distDir, fmt.Sprintf("%x.tar.gz", head))
	if err := archive.CreateDist(c, distFile); err != nil {
		return err
	}

	// 5. Save the signed head to ~/.config/ssotpub/pkgs/NAME/signed_head
	signedHead := filepath.Join(pkgDir, ssot.File)
	err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("%s: written\n", signedHead)

	// 6. Print the distribution name
	fmt.Println("")
	fmt.Printf("Please upload the following distribution file to: %s\n", URL)
	fmt.Println(distFile)
	fmt.Println("")

	// 7. Print DNS TXT records as defined by the .secpkg, the first signed head,
	//    and the URL.
	fmt.Println("Please publish the following DNS TXT records:")
	fmt.Println("")
	sh.TXTPrintHead(pkg.DNS)
	fmt.Println("")
	ssot.TXTPrintURL(pkg.DNS, URL)
	return nil
}

// CreatePkg implements the ssotpub 'createpkg' command.
func CreatePkg(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -name name -dns FQDN -url URL -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Create secure package and first signed head.\n")
		fs.PrintDefaults()
	}
	name := fs.String("name", "", "The project's package name")
	dns := fs.String("dns", "", "Fully qualified comain name for Codechain's TXT records (SSOT)")
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
