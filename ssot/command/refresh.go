package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/cloudflare"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func refresh(
	secpkgFilename string,
	validity time.Duration,
	secKeyRotate *[64]byte,
	sigRotate *[64]byte,
	commentRotate []byte,
) error {
	// 1. Parse the supplied .secpkg file.
	log.Println("1. parse .secpkg")
	pkg, err := secpkg.Load(secpkgFilename)
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

	// 4. Make sure the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
	//    matches the last signed HEAD in the .secpkg file.
	if prevSignedHead.Head() != pkg.Head {
		return fmt.Errorf("signed head in '%s' does not match HEAD in '%s'",
			signedHeadFile, secpkgFilename)
	}

	// 5. If ~/.config/ssotpub/pkgs/NAME/cloudflare.json exits, check the contained
	//    Cloudflare credentials and switch on automatic publishing of TXT records.
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

	// 6. If ROTATE is set, check if ~/.config/ssotput/pkgs/NAME/rotate_to exists.
	//    If it does, abort. Otherwise write public key to rotate to and rotate time
	//    to ~/.config/ssotput/pkgs/NAME/rotate_to.
	rotateToFile := filepath.Join(pkgDir, "rotate_to")
	log.Printf("6. if -rotate is set, check if '%s' exists", rotateToFile)
	if secKeyRotate != nil {
		log.Println("ROTATE set")
		exists, err := file.Exists(rotateToFile)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("-rotate set with existing file '%s'", rotateToFile)
		}
		err = ssot.WriteRotateTo(prevSignedHead, rotateToFile, secKeyRotate,
			sigRotate, commentRotate, validity)
		if err != nil {
			return err
		}
	} else {
		log.Println("ROTATE not set")
	}

	// 7. Create a new signed head with current HEAD, the counter of the previous
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
	log.Println("7. create a new signed head")
	pubKey := prevSignedHead.PubKey()
	var pubKeyRotate *[32]byte
	exists, err = file.Exists(rotateToFile)
	if err != nil {
		return err
	}
	var reached bool
	if exists {
		log.Println("rotate_to file exists")
		var rotateTo string
		rotateTo, reached, err = ssot.ReadRotateTo(rotateToFile)
		if err != nil {
			return err
		}
		if reached {
			log.Println("reached")
			pubKey = rotateTo
		} else {
			log.Println("set PUBKEY_ROTATE")
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

	newSignedHead, err := ssot.SignHead(prevSignedHead.HeadBuf(), prevSignedHead.Counter()+1,
		*secKey, pubKeyRotate, validity)
	if err != nil {
		return err
	}
	if err := ssot.RotateFile(newSignedHead, pkgDir); err != nil {
		return err
	}
	if reached {
		if err := os.Remove(rotateToFile); err != nil {
			return err
		}
	}

	// 8. Print DNS TXT record as defined by the .secpkg file and the signed head.
	//    If TXT record is to be published automatically, publish the TXT record.
	log.Println("8. print DNS TXT record")
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
	ssot.TXTPrintHead(newSignedHead, pkg.DNS)

	return nil
}

// Refresh implements the ssotpub 'refresh' command.
func Refresh(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s .secpkg [...]\n", argv0)
		fmt.Fprintf(os.Stderr, "Refresh head from .secpkg file(s).\n")
		fs.PrintDefaults()
	}
	rotate := fs.String("rotate", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	validity := fs.Duration("validity", ssot.MaximumValidity, "Validity of signed head")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	var (
		secKeyRotate  *[64]byte
		sigRotate     *[64]byte
		commentRotate []byte
	)
	if *rotate != "" {
		var err error
		secKeyRotate, sigRotate, commentRotate, err = seckey.Read(*rotate)
		if err != nil {
			return err
		}
	}
	for _, secpkgFilename := range fs.Args() {
		fmt.Printf("refreshing %s...\n", secpkgFilename)
		err := refresh(secpkgFilename, *validity, secKeyRotate, sigRotate,
			commentRotate)
		if err != nil {
			return err
		}
	}
	return nil
}
