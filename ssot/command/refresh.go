package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/dyn"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func refresh(secpkgFilename string) error {
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

	// 4. Make sure the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
	//    matches the HEAD in the .secpkg file.
	if prevSignedHead.Head() != pkg.Head {
		return fmt.Errorf("signed head in '%s' does not match HEAD in '%s'",
			signedHeadFile, secpkgFilename)
	}

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

	// 6. Create a new signed head with the same HEAD, the counter of the previous
	//    signed head plus 1, and update the saved signed head:
	//
	//    - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
	//             ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
	//    - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).
	log.Println("6. create a new signed head")
	newSignedHead := ssot.SignHead(prevSignedHead.HeadBuf(),
		prevSignedHead.Counter()+1, *secKey)
	if err := newSignedHead.RotateFile(pkgDir); err != nil {
		return err
	}

	// 7. Print DNS TXT record as defined by the .secpkg file and the signed head.
	//    If TXT records are to be published automatically, publish the TXT record.
	log.Println("7. print DNS TXT record")
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
	verbose := fs.Bool("v", false, "Be verbose")
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
	for _, secpkgFilename := range fs.Args() {
		fmt.Printf("refreshing %s...\n", secpkgFilename)
		if err := refresh(secpkgFilename); err != nil {
			return err
		}
	}
	return nil
}
