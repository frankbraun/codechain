package secpkg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
)

// Install pkg, see specification for details.
func (pkg *Package) Install() error {
	// 1. Has already been done by calling Load().

	// 2. Make sure the project has not been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME does not exist.
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", pkg.Name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("package already installed: '%s' exists", pkgDir)
	}

	// 3. Create directory ~/.config/secpkg/pkgs/NAME
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}

	// 4. Save .secpkg file to ~/.config/secpkg/pkgs/NAME/.secpkg
	fn := filepath.Join(pkgDir, File)
	err = ioutil.WriteFile(fn, []byte(pkg.Marshal()+"\n"), 0644)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	fmt.Printf("%s: written\n", File)

	// 5. Query TXT record from _codechain.DNS and validate the signed head
	//    contained in it (see ssot package).
	txts, err := net.LookupTXT(def.CodechainTXTName + pkg.DNS)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	var sh *ssot.SignedHead
	for _, txt := range txts {
		// parse TXT records and look for signed head
		sh, err = ssot.Unmarshal(txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot unmarshal: %s\n", txt)
			continue
		}
		fmt.Printf("signed head found: %s\n", sh.Head())
		break // TXT record found
	}
	if sh == nil {
		os.RemoveAll(pkgDir)
		return errors.New("secpkg: no valid TXT record found")
	}

	// 6. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head
	signedHead := filepath.Join(pkgDir, ssot.File)
	err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	fmt.Printf("%s: written\n", signedHead)

	// 7. Download distribution file from URL/HEAD_SSOT.tar.gz and save it to
	//    ~/.config/secpkg/pkgs/NAME/dists
	distDir := filepath.Join(pkgDir, "dists")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	fn = sh.Head() + ".tar.gz"
	filename := filepath.Join(distDir, fn)
	url := pkg.URL + "/" + fn
	fmt.Printf("download %s\n", url)
	err = file.Download(filename, url)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 8. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz
	//    to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
	//    -f ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz -head HEAD_SSOT`
	srcDir := filepath.Join(pkgDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := os.Chdir(srcDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	head := sh.HeadBuf()
	distFile := filepath.Join("..", "dists", fn)
	err = archive.ApplyFile(def.HashchainFile, def.PatchDir, distFile, &head)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := c.Close(); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := c.Apply(&head); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 9. Make sure HEAD_PKG is contained in
	//   ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain
	h, err := hex.Decode(pkg.Head, 32)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	copy(head[:], h)
	if err := c.CheckHead(head); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 10. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`
	buildDir := filepath.Join(pkgDir, "build")
	if err := file.CopyDir(srcDir, buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 11. Call `make prefix=~/.config/secpkg/local` in
	//     ~/.config/secpkg/pkgs/NAME/build
	localDir := filepath.Join(homedir.SecPkg(), "local")
	if err := os.Chdir(buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	prefix := fmt.Sprintf("prefix=%s", localDir)
	cmd := exec.Command("make", prefix)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 12. Call `make prefix=~/.config/secpkg/local install` in
	//     ~/.config/secpkg/pkgs/NAME/build
	cmd = exec.Command("make", prefix, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 13. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`
	installDir := filepath.Join(pkgDir, "install")
	if err := os.Rename(buildDir, installDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	return nil
}
