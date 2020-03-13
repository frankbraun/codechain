package command

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/cloudflare"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func writeTXTRecord(
	s *cloudflare.Session,
	zone string,
	DNS string,
	sh *ssot.SignedHead,
) error {
	// Update TXT record to publish the signed head.
	log.Println("update TXT record to publish the signed head")
	return s.TXTUpdate(zone, def.CodechainHeadName+DNS, sh.Marshal(), ssot.TTL)
}

func signHead(
	ctx context.Context,
	c *hashchain.HashChain,
	validity time.Duration,
	secKeyRotate *[64]byte,
	sigRotate *[64]byte,
	commentRotate []byte,
	secpkgFile string,
) error {
	// 1. Parse the .secpkg file in the current working directory.
	log.Println("1. parse .secpkg")
	pkg, err := secpkg.Load(secpkgFile)
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

	// 3. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head.
	log.Println("3. validate signed head")
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	prevSignedHead, err := ssot.Load(signedHeadFile)
	if err != nil {
		return err
	}

	// 4. Get the last signed HEAD from .codechain/hashchain in the current
	//    working directory.
	log.Println("4. get the head")
	head, _ := c.LastSignedHead()
	fmt.Printf("signing head %x\n", head)

	// 5. If ~/.config/ssotpub/pkgs/NAME/cloudflare.json exits, check the contained
	//    Cloudflare credentials and switch on automatic publishing of TXT records.
	log.Println("5. check Cloudflare credentials, if necessary")
	cloudflareFile := filepath.Join(pkgDir, cloudflare.ConfigFilename)
	exists, err = file.Exists(cloudflareFile)
	if err != nil {
		return err
	}
	var cloudflareSession *cloudflare.Session
	if exists {
		log.Printf("%s exists", cloudflareFile)
		cloudflareConfig, err := cloudflare.ReadConfig(cloudflareFile)
		if err != nil {
			return err
		}
		cloudflareSession, err = cloudflare.NewWithConfig(cloudflareConfig)
		if err != nil {
			return err
		}
	}

	// 6. Test build (see TestBuild specification).
	log.Println("6. test build")
	if err := testBuild(); err != nil {
		return err
	}

	// 7. If ROTATE is set, check if ~/.config/ssotput/pkgs/NAME/rotate_to exists.
	//    If it does, abort. Otherwise write public key to rotate to and rotate time
	//    to ~/.config/ssotput/pkgs/NAME/rotate_to.
	rotateToFile := filepath.Join(pkgDir, "rotate_to")
	log.Printf("7. if -rotate is set, check if '%s' exists", rotateToFile)
	if secKeyRotate != nil {
		exists, err := file.Exists(rotateToFile)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("-rotate set with existing file '%s'", rotateToFile)
		}
		err = prevSignedHead.WriteRotateTo(rotateToFile, secKeyRotate,
			sigRotate, commentRotate, validity)
		if err != nil {
			return err
		}
	}

	// 8. Create a new signed head with current HEAD, the counter of the previous
	//    signed head plus 1, and update the saved signed head:
	//
	//    - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
	//             ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
	//    - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).
	//
	//    If ~/.config/ssotput/pkgs/NAME/rotate_to exists:
	//
	//    - If rotate time has been reached use pubkey from file as PUBKEY and
	//      remove ~/.config/ssotput/pkgs/NAME/rotate_to.
	//    - Otherwise use old PUBKEY and set pubkey from file as PUBKEY_ROTATE.
	log.Println("8. create a new signed head")
	pubKey := prevSignedHead.PubKey()
	var pubKeyRotate *[32]byte
	exists, err = file.Exists(rotateToFile)
	if err != nil {
		return err
	}
	var reached bool
	if exists {
		var rotateTo string
		rotateTo, reached, err = ssot.ReadRotateTo(rotateToFile)
		if err != nil {
			return err
		}
		if reached {
			pubKey = rotateTo
		} else {
			pk, err := base64.Decode(rotateTo, 32)
			if err != nil {

			}
			pubKeyRotate = new([32]byte)
			copy(pubKeyRotate[:], pk)
		}
	}

	secKeyFile := filepath.Join(homedir.SSOTPub(), def.SecretsSubDir, pubKey)
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}

	newSignedHead, err := ssot.SignHead(head, prevSignedHead.Counter()+1,
		*secKey, pubKeyRotate, validity)
	if err != nil {
		return err
	}
	if err := newSignedHead.RotateFile(pkgDir); err != nil {
		return err
	}
	if reached {
		if err := os.Remove(rotateToFile); err != nil {
			return err
		}
	}

	// 9. If the HEAD changed, save the current distribution to:
	//    ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`).
	log.Println("9. if the HEAD changed, save the current distribution")
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

	// 10. If the HEAD changed, lookup the download URLs and print where to upload
	//     the distribution file:
	//     ~/.config/ssotpkg/pkgs/NAME/dists/HEAD.tar.gz
	log.Println("10. if the HEAD changed, lookup the download URLs")
	if h != pkg.Head {
		URLs, err := ssot.LookupURLs(ctx, pkg.DNS)
		if err != nil {
			return err
		}
		fmt.Println("")
		fmt.Println("Please upload the following distribution file to:")
		for _, URL := range URLs {
			fmt.Println(URL)
		}
		fmt.Println("")
		fmt.Println(distFile)
		fmt.Println("")
	}

	// 11. Print DNS TXT record as defined by the .secpkg and the signed head.
	//     If TXT records are to be published automatically, publish the TXT record.
	log.Println("11. print DNS TXT record")
	if cloudflareSession != nil {
		// Write TXT record
		log.Printf("DNS=%s", pkg.DNS)
		parts := strings.Split(pkg.DNS, ".")
		zone := parts[len(parts)-2] + "." + parts[len(parts)-1]
		err := writeTXTRecord(cloudflareSession, zone, pkg.DNS, newSignedHead)
		if err != nil {
			return nil
		}
		fmt.Println("The following DNS TXT record has been published:")
	} else {
		fmt.Println("Please publish the following DNS TXT record:")
	}
	fmt.Println("")
	newSignedHead.TXTPrintHead(pkg.DNS)

	// 12. If the HEAD changed, update the .secpkg file accordingly.
	log.Println("12. if the HEAD changed, update the .secpkg file")
	if h != pkg.Head {
		pkg.Head = h
		newSecPkgFile := secpkgFile + "_new"
		err = ioutil.WriteFile(newSecPkgFile, []byte(pkg.Marshal()+"\n"), 0644)
		if err != nil {
			return err
		}
		if err := os.Rename(newSecPkgFile, secpkgFile); err != nil {
			return err
		}
		fmt.Printf("\n%s: updated\n", secpkgFile)
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
	secpkgFile := fs.String("f", secpkg.File, "The secpkg filename")
	rotate := fs.String("rotate", "", "Secret key file")
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
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	var (
		secKeyRotate  *[64]byte
		sigRotate     *[64]byte
		commentRotate []byte
	)
	if *rotate != "" {
		secKeyRotate, sigRotate, commentRotate, err = seckey.Read(*rotate)
		if err != nil {
			return err
		}
	}
	// add interrupt handler
	interrupt.AddInterruptHandler(func() {
		c.Close()
	})
	// run signHead
	go func() {
		err := signHead(context.Background(), c, *validity, secKeyRotate,
			sigRotate, commentRotate, *secpkgFile)
		if err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
