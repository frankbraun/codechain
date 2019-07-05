package secpkg

import (
	"fmt"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
)

func checkUpdate(visited map[string]bool, name string) (bool, error) {
	// 1. Make sure the project with NAME has been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME exists.
	//    Set SKIP_CHECK and NEEDS_UPDATE to false.
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return false, err
	}
	if !exists {
		return false,
			fmt.Errorf("package not installed: '%s' does not exist", pkgDir)
	}
	skipCheck := false
	needsUpdate := false

	// 2. Load .secpkg file from ~/.config/secpkg/pkgs/NAME/.secpkg
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
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	shDisk, err := ssot.Load(signedHeadFile)
	if err != nil {
		return false, err
	}

	// 4. Query TXT record from _codechain-head.DNS, if it is the same as DISK, set
	//    SKIP_CHECK to true.
	shDNS, err := ssot.LookupHead(pkg.DNS)
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
	}

	// 6. If not SKIP_CHECK and if signed head from TXT record not the same as the
	//    one from DISK, set SKIP_CHECK and NEEDS_UPDATE to true.
	if !skipCheck {
		if shDNS.Head() == shDisk.Head() {
			skipCheck = true
			needsUpdate = true
		}
	}

	// 7. If not SKIP_CHECK, check if HEAD is contained in
	//    ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain.
	//    If not, set NEEDS_UPDATE to true.
	if !skipCheck {
		c, err := hashchain.ReadFile(def.HashchainFile)
		if err != nil {
			return false, err
		}
		if err := c.Close(); err != nil {
			return false, err
		}
		if err := c.CheckHead(shDNS.HeadBuf()); err != nil {
			needsUpdate = true
		}
	}

	// 8. If NEEDS_UPDATE is false, check if the directory
	//    ~/.config/secpkg/pkgs/NAME/src/.secdep exists and contains any .secpkg
	//    files, ensure these secure dependencies are installed and up-to-date. If
	//    at least one dependency needs an update, set NEEDS_UPDATE to true.
	if !needsUpdate {
		needsUpdate, err = ensureCheckUpdate(visited, name)
		if err != nil {
			return false, err
		}
	}

	// 9. Update signed head:
	//
	//    - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
	//             ~/.config/secpkg/pkgs/NAME/previous_signed_head`
	//    - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
	if err := shDNS.RotateFile(pkgDir); err != nil {
		return false, nil
	}

	// 10. Return NEEDS_UPDATE.
	return needsUpdate, nil
}

// CheckUpdate package with name, see specification for details.
func CheckUpdate(name string) (bool, error) {
	visited := make(map[string]bool)
	visited[name] = true
	return checkUpdate(visited, name)
}
