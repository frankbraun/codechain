package secpkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/gnumake"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/homedir"
)

func (pkg *Package) install(ctx context.Context, visited map[string]bool) error {
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
	fmt.Printf("%s: written\n", fn)

	// 5. Query TXT record from _codechain-head.DNS and validate the signed head
	//    contained in it (see ssot package).
	sh, err := ssot.LookupHead(ctx, pkg.DNS)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 6. Query TXT record from _codechain-url.DNS and save it as URL.
	URL, err := ssot.LookupURL(ctx, pkg.DNS)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 7. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head
	signedHead := filepath.Join(pkgDir, ssot.File)
	err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	fmt.Printf("%s: written\n", signedHead)

	// 8. Download distribution file from URL/HEAD_SSOT.tar.gz and save it to
	//    ~/.config/secpkg/pkgs/NAME/dists
	distDir := filepath.Join(pkgDir, "dists")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	var encSuffix string
	if pkg.Key != "" {
		encSuffix = ".enc"
	}
	fn = sh.Head() + ".tar.gz" + encSuffix
	filename := filepath.Join(distDir, fn)
	url := URL + "/" + fn
	fmt.Printf("download %s\n", url)
	err = file.Download(filename, url)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 9. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz
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
	if pkg.Key != "" {
		key, err := pkg.GetKey()
		if err != nil {
			return err
		}
		err = archive.ApplyEncryptedFile(def.HashchainFile, def.PatchDir,
			distFile, &head, key)
		if err != nil {
			os.RemoveAll(pkgDir)
			return err
		}
	} else {
		err = archive.ApplyFile(def.HashchainFile, def.PatchDir, distFile, &head)
		if err != nil {
			os.RemoveAll(pkgDir)
			return err
		}
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

	// 10. Make sure HEAD_PKG is contained in
	//     ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain
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

	// 11. If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
	//     contains any .secpkg files, ensure these secure dependencies are
	//     installed and up-to-date.
	if _, err := ensure(ctx, visited, pkg.Name); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 12. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`
	buildDir := filepath.Join(pkgDir, "build")
	if err := file.CopyDir(srcDir, buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 13. Call `make prefix=~/.config/secpkg/local` in
	//     ~/.config/secpkg/pkgs/NAME/build
	localDir := filepath.Join(homedir.SecPkg(), "local")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := os.Chdir(buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	/* TODO: call via $SHELL
	shell := os.Getenv("SHELL")
	if shell == "" {
		os.RemoveAll(pkgDir)
		return errors.New("secpkg: $SHELL not defined")
	}
	*/
	if err := gnumake.Call(localDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 14. Call `make prefix=~/.config/secpkg/local install` in
	//     ~/.config/secpkg/pkgs/NAME/build
	if err := gnumake.Install(localDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 15. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`
	installedDir := filepath.Join(pkgDir, "installed")
	if err := os.Rename(buildDir, installedDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	return nil
}

// Install pkg, see specification for details.
func (pkg *Package) Install(ctx context.Context) error {
	visited := make(map[string]bool)
	visited[pkg.Name] = true
	return pkg.install(ctx, visited)
}
