package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/dyn"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func writeTXTRecord(
	s *dyn.Session,
	zone string,
	DNS string,
	sh *ssot.SignedHead,
) error {
	// Update TXT record to publish the signed head.
	log.Println("update TXT record to publish the signed head")
	err := s.TXTUpdate(zone, def.CodechainHeadName+DNS, sh.Marshal(), ssot.TTL)
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
	return s.ZoneUpdate(zone)
}

func signHead(c *hashchain.HashChain, validity time.Duration) error {
	// 1. Parse the .secpkg file in the current working directory.
	log.Println("1. parse .secpkg")
	pkg, err := secpkg.Load(secpkg.File)
	if err != nil {
		return err
	}

	// 2. Make sure the project with NAME has been published before.
	//    That is, the directory ~/.config/ssotpub/pkgs/NAME exists.
	log.Println("2. make sure project exists")
	pkgDir := filepath.Join(homedir.SSOTPub(), "pkgs", pkg.Name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("package not published yet: '%s' does not exist", pkgDir)
	}

	// 3. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
	//    and make sure the corresponding secret key is available.
	log.Println("3. validate signed head")
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	prevSignedHead, err := ssot.Load(signedHeadFile)
	if err != nil {
		return err
	}

	secKeyFile := filepath.Join(homedir.SSOTPub(), def.SecretsSubDir, prevSignedHead.PubKey())
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}

	// 4. Get the HEAD from .codechain/hashchain in the current working directory.
	log.Println("4. get the head")
	head := c.Head()
	fmt.Printf("signing head %x\n", head)

	// 5. If ~/.config/ssotpub/pkgs/NAME/dyn.json exits, check the contained Dyn
	//    credentials and switch on automatic publishing of TXT records.
	dynFile := filepath.Join(pkgDir, dyn.ConfigFilename)
	exists, err = file.Exists(dynFile)
	if err != nil {
		return err
	}
	var dynSession *dyn.Session
	if exists {
		log.Printf("%s exists", dynFile)
		dynConfig, err := dyn.ReadConfig(dynFile)
		if err != nil {
			return err
		}
		dynSession, err = dyn.NewWithConfig(dynConfig)
		if err != nil {
			return err
		}
		defer dynSession.Close()
	}

	// 6. Create a new signed head with current HEAD, the counter of the previous
	//    signed head plus 1, and update the saved signed head:
	//
	//    - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
	//           ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
	//    - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).
	log.Println("6. create a new signed head")
	newSignedHead, err := ssot.SignHead(head, prevSignedHead.Counter()+1,
		*secKey, validity)
	if err != nil {
		return err
	}
	if err := newSignedHead.RotateFile(pkgDir); err != nil {
		return err
	}

	// 7. If the HEAD changed, save the current distribution to:
	//    ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`).
	log.Println("7. if the HEAD changed, save the current distribution")
	h := hex.Encode(head[:])
	var distFile string
	if h != pkg.Head {
		var encSuffix string
		if pkg.Key != "" {
			encSuffix = ".enc"
		}
		distDir := filepath.Join(pkgDir, "dists")
		distFile = filepath.Join(distDir, fmt.Sprintf("%x.tar.gz%s", head, encSuffix))
		if pkg.Key != "" {
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
	}

	// 8. If the HEAD changed, lookup the download URL and print where to upload
	//    the distribution file:
	//    ~/.config/ssotpkg/pkgs/NAME/dists/HEAD.tar.gz
	log.Println("8. if the HEAD changed, lookup the download URL")
	if h != pkg.Head {
		URL, err := ssot.LookupURL(pkg.DNS)
		if err != nil {
			return err
		}
		fmt.Println("")
		fmt.Printf("Please upload the following distribution file to: %s\n", URL)
		fmt.Println(distFile)
		fmt.Println("")
	}

	// 9. Print DNS TXT record as defined by the .secpkg and the signed head.
	// If TXT records are to be published automatically, publish the TXT record.
	log.Println("9. print DNS TXT record")
	if dynSession != nil {
		// Write TXT record
		log.Printf("DNS=%s", pkg.DNS)
		parts := strings.Split(pkg.DNS, ".")
		zone := parts[len(parts)-2] + "." + parts[len(parts)-1]
		err := writeTXTRecord(dynSession, zone, pkg.DNS, newSignedHead)
		if err != nil {
			return nil
		}
		fmt.Println("The following DNS TXT record has been published:")
	} else {
		fmt.Println("Please publish the following DNS TXT record:")
	}
	fmt.Println("")
	newSignedHead.TXTPrintHead(pkg.DNS)

	// 10. If the HEAD changed, update the .secpkg file accordingly.
	log.Println("10. if the HEAD changed, update the .secpkg file")
	if h != pkg.Head {
		pkg.Head = h
		newSecPkgFile := secpkg.File + "_new"
		err = ioutil.WriteFile(newSecPkgFile, []byte(pkg.Marshal()+"\n"), 0644)
		if err != nil {
			return err
		}
		if err := os.Rename(newSecPkgFile, secpkg.File); err != nil {
			return err
		}
		fmt.Printf("\n%s: updated\n", secpkg.File)
	}

	return nil
}

// SignHead implements the ssotpub 'signhead' command.
func SignHead(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Sign Codechain head and print it on stdout.\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Be verbose")
	validity := fs.Duration("validity", ssot.MaximumValidity, "Validity of signed head")
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
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	// add interrupt handler
	interrupt.AddInterruptHandler(func() {
		c.Close()
	})
	// run signHead
	go func() {
		if err := signHead(c, *validity); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
