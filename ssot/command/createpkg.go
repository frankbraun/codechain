package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/cloudflare"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func writeTXTRecords(
	s *cloudflare.Session,
	zone string,
	DNS string,
	sh *ssot.SignedHead,
	URL string,
) error {
	// Create TXT record to publish the signed head.
	log.Println("create TXT record to publish the signed head")
	_, err := s.TXTCreate(zone, def.CodechainHeadName+DNS, sh.Marshal(), ssot.TTL)
	if err != nil {
		return err
	}
	// Create TXT record to publish the url.
	log.Println("create TXT record to publish the url")
	jsn, err := s.TXTCreate(zone, def.CodechainURLName+DNS, URL, ssot.TTL)
	if err != nil {
		return err
	}
	log.Println(jsn)
	return nil
}

func createPkg(
	c *hashchain.HashChain, name, dns, URL, secKeyFile, secpkgFile string,
	encrypted, useCloudflare bool,
	apiKey, email string,
	validity time.Duration,
) error {
	head := c.Head()
	fmt.Printf("create package for head %x\n", head)
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}
	// 4. Create package (before 1., because this checks the arguments)
	if _, err := url.Parse(URL); err != nil {
		return err
	}
	pkg, err := secpkg.New(name, dns, head, encrypted)
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
		cloudflareConfig  *cloudflare.Config
		cloudflareSession *cloudflare.Session
	)
	if useCloudflare {
		cloudflareConfig = &cloudflare.Config{
			APIKey: apiKey,
			Email:  email,
		}
		cloudflareSession, err = cloudflare.NewWithConfig(cloudflareConfig)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Publishing TXT records manually, restart with -cloudflare to switch to automatic")
	}

	// 3. Test build (see TestBuild specification).
	fmt.Println("call testBuild()")
	if err := testBuild(); err != nil {
		return err
	}
	fmt.Println("done testBuild()")

	// Create .secpkg file
	exists, err = file.Exists(secpkgFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("secure package already exists: %s", secpkgFile)
	}
	err = ioutil.WriteFile(secpkgFile, []byte(pkg.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("%s: written\n", secpkgFile)

	// 5. Create the first signed head with counter set to 0.
	sh, err := ssot.SignHead(head, 0, *secKey, nil, validity)
	if err != nil {
		return err
	}

	// 6. Create the directory ~/.config/ssotpub/pkgs/NAME/dists
	//    and save the current distribution to
	//    ~/.config/ssotpub/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`)
	distDir := filepath.Join(pkgDir, "dists")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	var encSuffix string
	if encrypted {
		encSuffix = ".enc"
	}
	distFile := filepath.Join(distDir, fmt.Sprintf("%x.tar.gz%s", head, encSuffix))
	if encrypted {
		key, err := pkg.GetKey()
		if err != nil {
			return err
		}
		if err := archive.CreateEncryptedDist(c, distFile, key); err != nil {
			return err
		}
	} else {
		if err := archive.CreateDist(c, distFile); err != nil {
			return err
		}
	}

	// 7. Save the signed head to ~/.config/ssotpub/pkgs/NAME/signed_head
	signedHead := filepath.Join(pkgDir, ssot.File)
	err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("%s: written\n", signedHead)

	// 8. Print the distribution name
	fmt.Println("")
	fmt.Printf("Please upload the following distribution file to: %s\n", URL)
	fmt.Println(distFile)
	fmt.Println("")

	// 9. Print DNS TXT records as defined by the .secpkg, the first signed head,
	//    and the download URL. If TXT records are to be published automatically,
	//    save credentials and publish the TXT record.
	if useCloudflare {
		// Save the credentials to ~/.config/ssotpub/pkgs/NAME/cloudflare.json
		cloudflareFile := filepath.Join(pkgDir, cloudflare.ConfigFilename)
		if err := cloudflareConfig.Write(cloudflareFile); err != nil {
			return err
		}
		// Write TXT records
		log.Printf("DNS=%s", pkg.DNS)
		parts := strings.Split(pkg.DNS, ".")
		zone := parts[len(parts)-2] + "." + parts[len(parts)-1]
		err := writeTXTRecords(cloudflareSession, zone, pkg.DNS, sh, URL)
		if err != nil {
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
	secpkgFile := fs.String("f", secpkg.File, "The secpkg filename")
	name := fs.String("name", "", "The project's package name")
	dns := fs.String("dns", "", "Fully qualified comain name for Codechain's TXT records (SSOT)")
	url := fs.String("url", "", "URL to download project files from (URL/head.tar.gz)")
	secKey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	encrypted := fs.Bool("encrypted", false, "Encrypt source code archives")
	useCloudflare := fs.Bool("cloudflare", false, "Use Cloudflare API to publish TXT records automatically")
	apiKey := fs.String("api-key", "", "Cloudflare API key")
	email := fs.String("email", "", "Email address associated with Cloudflare account")
	validity := fs.Duration("validity", ssot.MaximumValidity, "Validity of signed head")
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
	if *useCloudflare && *apiKey == "" {
		return fmt.Errorf("%s: option -cloudflare requires option -api-key", argv0)
	}
	if *useCloudflare && *email == "" {
		return fmt.Errorf("%s: option -cloudflare requires option -email", argv0)
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
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
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
		err := createPkg(c, *name, *dns, *url, *secKey, *secpkgFile, *encrypted,
			*useCloudflare, *apiKey, *email, *validity)
		if err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
