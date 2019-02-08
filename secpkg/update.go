package secpkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/gnumake"
	"github.com/frankbraun/codechain/util/homedir"
)

// Update package with name, see specification for details.
func Update(name string) error {
	// 1. Make sure the project with NAME has been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME exists.
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("package not installed: '%s' does not exist", pkgDir)
	}

	// 2. Load .secpkg file from ~/.config/secpkg/pkgs/NAME/.secpkg
	fn := filepath.Join(pkgDir, File)
	pkg, err := Load(fn)
	if err != nil {
		return err
	}
	if pkg.Name != name {
		return fmt.Errorf("package to update (%s) differs from package name in %s", name, fn)
	}

	// 3. Load signed head from ~/.config/secpkg/pkgs/NAME/signed_head (as DISK)
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	shDisk, err := ssot.Load(signedHeadFile)
	if err != nil {
		return err
	}

	// 4. Query TXT record from _codechain-head.DNS, if it is the same as DISK, goto 16.
	shDNS, err := ssot.LookupHead(pkg.DNS)
	if err != nil {
		return err
	}
	if shDisk.Marshal() == shDNS.Marshal() {
		fmt.Printf("package '%s' already up-to-date\n", name)
		return nil
	}

	// 5. Query TXT record from _codechain-url.DNS and save it as URL.
	URL, err := ssot.LookupURL(pkg.DNS)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 6. Validate signed head from TXT (also see ssot package) and store HEAD:
	//
	//    - pubKey from TXT must be the same as pubKey or pubKeyRotate from DISK.
	//    - The counter from TXT must be larger than the counter from DISK.
	//    - The signed head must be valid (as defined by validFrom and validTo).
	//
	// If the validation fails, abort update procedure and report error.
	if !(shDNS.PubKey() == shDisk.PubKey() ||
		shDNS.PubKey() == shDisk.PubKeyRotate()) {
		return fmt.Errorf("secpkg: public key from TXT record does not match public key (or rotate) from disk")
	}
	if shDNS.Counter() <= shDisk.Counter() {
		return fmt.Errorf("secpkg: counter from TXT record is not increasing")
	}
	if err := shDNS.Valid(); err != nil {
		return err
	}

	// 7. If signed head from TXT record is the same as the one from DISK:
	//
	//    - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
	//             ~/.config/secpkg/pkgs/NAME/previous_signed_head`
	//     - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
	//     - Goto 16.

	if shDNS.Head() == shDisk.Head() {
		return shDNS.RotateFile(pkgDir)
	}

	// 8. Download distribution file from URL/HEAD.tar.gz and save it to
	//    ~/.config/secpkg/pkgs/NAME/dists
	distDir := filepath.Join(pkgDir, "dists")
	var encSuffix string
	if pkg.Key != "" {
		encSuffix = ".enc"
	}
	fn = shDNS.Head() + ".tar.gz" + encSuffix
	filename := filepath.Join(distDir, fn)
	url := URL + "/" + fn
	fmt.Printf("download %s\n", url)
	err = file.Download(filename, url)
	if err != nil {
		return err
	}

	// 9. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz
	//	  to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
	//	  -f ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz -head HEAD`.
	srcDir := filepath.Join(pkgDir, "src")
	if err := os.Chdir(srcDir); err != nil {
		return err
	}
	head := shDNS.HeadBuf()
	distFile := filepath.Join("..", "dists", fn)
	if pkg.Key != "" {
		key, err := pkg.GetKey()
		if err != nil {
			return err
		}
		err = archive.ApplyEncryptedFile(def.HashchainFile, def.PatchDir,
			distFile, &head, key)
		if err != nil {
			return err
		}
	} else {
		err = archive.ApplyFile(def.HashchainFile, def.PatchDir, distFile, &head)
		if err != nil {
			return err
		}
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}
	if err := c.Apply(&head); err != nil {
		return err
	}

	// 10. `rm -rf ~/.config/secpkg/pkgs/NAME/build`
	buildDir := filepath.Join(pkgDir, "build")
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}

	// 11. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`
	if err := file.CopyDir(srcDir, buildDir); err != nil {
		return err
	}

	// 12. Call `make prefix=~/.config/secpkg/local` in
	//     ~/.config/secpkg/pkgs/NAME/build
	localDir := filepath.Join(homedir.SecPkg(), "local")
	if err := os.Chdir(buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := gnumake.Call(localDir); err != nil {
		return err
	}

	// 13. Call `make prefix= ~/.config/secpkg/local install` in
	//     ~/.config/secpkg/pkgs/NAME/build
	if err := gnumake.Install(localDir); err != nil {
		return err
	}

	// 14. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`
	installedDir := filepath.Join(pkgDir, "installed")
	if err := os.RemoveAll(installedDir); err != nil {
		return err
	}

	if err := os.Rename(buildDir, installedDir); err != nil {
		return err
	}

	// 15. Update signed head:
	//
	//      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
	//               ~/.config/secpkg/pkgs/NAME/previous_signed_head`
	//      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
	return shDNS.RotateFile(pkgDir)

	// 16. The software has been successfully updated.
}
