package secpkg

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

func checkUpdate(ctx context.Context, visited map[string]bool, name string) (bool, error) {
	// 1. Make sure the project with NAME has been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME exists.
	//    Set SKIP_CHECK and NEEDS_UPDATE to false.
	log.Printf("1. make sure '%s' has been installed\n", name)
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrNotInstalled
	}
	skipCheck := false
	needsUpdate := false

	// 2. Load .secpkg file from ~/.config/secpkg/pkgs/NAME/.secpkg
	log.Println("2. load .secpkg file")
	fn := filepath.Join(pkgDir, File)
	pkg, err := Load(fn)
	if err != nil {
		return false, err
	}
	if pkg.Name != name {
		return false,
			fmt.Errorf("package to update (%s) differs from package name in %s", name, fn)
	}

	// 3. Load signed head from ~/.config/secpkg/pkgs/NAME/signed_head (as DISK)
	log.Println("3. load signed head")
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	shDisk, err := ssot.Load(signedHeadFile)
	if err != nil {
		return false, err
	}

	// 4. Query TXT record from _codechain-head.DNS, if it is the same as DISK, set
	//    SKIP_CHECK to true.
	log.Println("4. query TX record")
	shDNS, err := ssot.LookupHead(ctx, pkg.DNS)
	if err != nil {
		return false, err
	}
	if shDisk.Marshal() == shDNS.Marshal() {
		skipCheck = true
	}

	// 5. If not SKIP_CHECK, validate signed head from TXT (also see ssot package)
	//    and store HEAD:
	//
	// 	  - pubKey from TXT must be the same as pubKey or pubKeyRotate from DISK,
	// 	    if the signed head from DISK is not expired.
	// 	  - The counter from TXT must be larger than the counter from DISK.
	// 	  - The signed head must be valid (as defined by validFrom and validTo).
	//
	// If the validation fails, abort check update procedure and report error.
	log.Println("5. validate signed head")
	if !skipCheck {
		if err := shDisk.Valid(); err == nil { // not expired
			if !(shDNS.PubKey() == shDisk.PubKey() ||
				shDNS.PubKey() == shDisk.PubKeyRotate()) {
				return false,
					fmt.Errorf("secpkg: public key from TXT record does not match public key (or rotate) from disk")
			}
		}
		if shDNS.Counter() <= shDisk.Counter() {
			return false,
				fmt.Errorf("secpkg: counter from TXT record is not increasing")
		}
		if err := shDNS.Valid(); err != nil {
			return false, err
		}
	} else {
		log.Println("skipped")
	}

	// 6. If not SKIP_CHECK and if signed head from TXT record not the same as the
	//    one from DISK, set SKIP_CHECK and NEEDS_UPDATE to true.
	log.Println("6. compare signed heads")
	if !skipCheck {
		if shDNS.Head() == shDisk.Head() {
			log.Println("set SKIP_CHECK and NEEDS_UPDATE to true")
			skipCheck = true
			needsUpdate = true
		}
	} else {
		log.Println("skipped")
	}

	// 7. If not NEEDS_UPDATE, check if HEAD is contained in
	//    ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain.
	//    If not, set NEEDS_UPDATE to true.
	log.Println("7. check if HEAD is contained in hashchain")
	if !needsUpdate {
		srcDir := filepath.Join(pkgDir, "src")
		c, err := hashchain.ReadFile(filepath.Join(srcDir, def.HashchainFile))
		if err != nil {
			return false, err
		}
		if err := c.Close(); err != nil {
			return false, err
		}
		log.Printf("c.CheckHead(%s)\n", shDNS.Head())
		if err := c.CheckHead(shDNS.HeadBuf()); err != nil {
			log.Println("set NEEDS_UPDATE=true")
			needsUpdate = true
		}
	} else {
		log.Println("skipped")
	}

	// 8. If NEEDS_UPDATE is false, check if the directory
	//    ~/.config/secpkg/pkgs/NAME/src/.secdep exists and contains any .secpkg
	//    files, ensure these secure dependencies are installed and up-to-date. If
	//    at least one dependency needs an update, set NEEDS_UPDATE to true.
	log.Println("8. check .secdep directory")
	if !needsUpdate {
		needsUpdate, err = ensureCheckUpdate(ctx, visited, name)
		if err != nil {
			return false, err
		}
	} else {
		log.Println("skipped")
	}

	// 9. Update signed head:
	//
	//    - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
	//             ~/.config/secpkg/pkgs/NAME/previous_signed_head`
	//    - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
	log.Println("9. update signed head")
	if err := shDNS.RotateFile(pkgDir); err != nil {
		return false, nil
	}

	// 10. Return NEEDS_UPDATE.
	log.Printf("10. return NEEDS_UPDATE=%s\n", strconv.FormatBool(needsUpdate))
	return needsUpdate, nil
}

// CheckUpdate checks installed package with name for updates, see
// specification for details.
func CheckUpdate(ctx context.Context, name string) (bool, error) {
	visited := make(map[string]bool)
	visited[name] = true
	return checkUpdate(ctx, visited, name)
}
