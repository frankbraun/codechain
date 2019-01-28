package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/dyn"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func writeTXTRecords(
	s *dyn.Session,
	zone string,
	pkg *secpkg.Package,
	sh *ssot.SignedHead,
	URL string,
) error {
	// Create TXT record to publish the signed head.
	log.Println("create TXT record to publish the signed head")
	err := s.TXTCreate(zone, def.CodechainHeadName+pkg.DNS, sh.Marshal(), ssot.TTL)
	if err != nil {
		return err
	}
	// Create TXT record to publish the url.
	log.Println("create TXT record to publish the url")
	err = s.TXTCreate(zone, def.CodechainURLName+pkg.DNS, URL, ssot.TTL)
	if err != nil {
		return err
	}
	ret, err := s.ZoneChangeset(zone)
	if err != nil {
		return err
	}
	jsn, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		return err
	}
	log.Println(string(jsn))
	if err := s.ZoneUpdate(zone); err != nil {
		return err
	}
	return nil
}

func createPkg(
	c *hashchain.HashChain, name, dns, URL, secKeyFile string,
	useDyn bool,
	customerName, userName, password string,
) error {
	head := c.Head()
	fmt.Printf("create package for head %x\n", head)
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}
	// 3. Create package (before 1., because this checks the arguments)
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

	// 2. If TXT records are to be published automatically, check credentials.
	var (
		dynConfig  *dyn.Config
		dynSession *dyn.Session
	)
	if useDyn {
		dynConfig = &dyn.Config{
			CustomerName: customerName,
			UserName:     userName,
			Password:     password,
		}
		dynSession, err = dyn.NewWithConfig(dynConfig)
		if err != nil {
			return err
		}
		defer dynSession.Close()
	} else {
		fmt.Println("Publishing TXT records manually, restart with -dyn to switch to automatic")
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

	// 4. Create the first signed head with counter set to 0.
	sh := ssot.SignHead(head, 0, *secKey)

	// 5. Create the directory ~/.config/ssotpub/pkgs/NAME/dists
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

	// 6. Save the signed head to ~/.config/ssotpub/pkgs/NAME/signed_head
	signedHead := filepath.Join(pkgDir, ssot.File)
	err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("%s: written\n", signedHead)

	// 7. Print the distribution name
	fmt.Println("")
	fmt.Printf("Please upload the following distribution file to: %s\n", URL)
	fmt.Println(distFile)
	fmt.Println("")

	// 8. Print DNS TXT records as defined by the .secpkg, the first signed head,
	//    and the download URL. If TXT records are to be published automatically,
	//    save credentials and publish the TXT record.
	if useDyn {
		// Save the credentials to ~/.config/ssotpub/pkgs/NAME/dyn.json
		dynFile := filepath.Join(pkgDir, dyn.ConfigFilename)
		if err := dynConfig.Write(dynFile); err != nil {
			return err
		}
		// Write TXT records
		log.Printf("dns=%s", dns)
		parts := strings.Split(dns, ".")
		zone := parts[len(parts)-2] + "." + parts[len(parts)-1]
		if err := writeTXTRecords(dynSession, zone, pkg, sh, URL); err != nil {
			return nil
		}
		fmt.Println("The following DNS TXT records have been published:")
	} else {
		fmt.Println("Please publish the following DNS TXT records:")
	}
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
	useDyn := fs.Bool("dyn", false, "Use Dyn Managed DNS API to publish TXT records automatically")
	customerName := fs.String("customer", "", "Customer name for Dyn Managed DNS API")
	userName := fs.String("user", "", "User name for Dyn Managed DNS API")
	password := fs.String("password", "", "Password for Dyn Managed DNS API")
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
	if *useDyn && *customerName == "" {
		return fmt.Errorf("%s: option -dyn requires option -customer", argv0)
	}
	if *useDyn && *userName == "" {
		return fmt.Errorf("%s: option -dyn requires option -user", argv0)
	}
	if *useDyn && *password == "" {
		return fmt.Errorf("%s: option -dyn requires option -password", argv0)
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
		err := createPkg(c, *name, *dns, *url, *secKey, *useDyn, *customerName,
			*userName, *password)
		if err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
