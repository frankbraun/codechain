package secpkg

import (
	"context"
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

func update(ctx context.Context, visited map[string]bool, name string) (bool, error) {
	// 1. Make sure the project with NAME has been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME exists.
	//    Set SKIP_BUILD to false.
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrNotInstalled
	}
	skipBuild := false

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
	//    SKIP_BUILD to true.
	shDNS, err := ssot.LookupHead(ctx, pkg.DNS)
	if err != nil {
		return false, err
	}
	if shDisk.Marshal() == shDNS.Marshal() {
		skipBuild = true
	}

	// 5. Query TXT record from _codechain-url.DNS and save it as URL.
	URL, err := ssot.LookupURL(ctx, pkg.DNS)
	if err != nil {
		os.RemoveAll(pkgDir)
		return false, err
	}

	// 6. If not SKIP_BUILD, validate signed head from TXT (also see ssot package)
	//    and store HEAD:
	//
	//    - pubKey from TXT must be the same as pubKey or pubKeyRotate from DISK
	//      if the signed head from DISK is not expired.
	//    - The counter from TXT must be larger than the counter from DISK.
	//    - The signed head must be valid (as defined by validFrom and validTo).
	//
	//    If the validation fails, abort update procedure and report error.
	if !skipBuild {
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

	// 7. If not SKIP_BUILD and if signed head from TXT record is the same as the
	//    one from DISK, set SKIP_BUILD to true.
	if !skipBuild {
		if shDNS.Head() == shDisk.Head() {
			skipBuild = true
		}
	}

	// 8. If SKIP_BUILD, check if HEAD is contained in
	//    ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain.
	//    If not, set SKIP_BUILD to false.
	//    This can happend if we checked for updates.
	srcDir := filepath.Join(pkgDir, "src")
	if skipBuild {
		c, err := hashchain.ReadFile(filepath.Join(srcDir, def.HashchainFile))
		if err != nil {
			return false, err
		}
		if err := c.Close(); err != nil {
			return false, err
		}
		if err := c.CheckHead(shDNS.HeadBuf()); err != nil {
			skipBuild = false
		}
	}

	// 9. If not SKIP_BUILD, download distribution file from URL/HEAD.tar.gz and
	//    save it to ~/.config/secpkg/pkgs/NAME/dists
	if !skipBuild {
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
			return false, err
		}
	}

	// 10. If not SKIP_BUILD, apply ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz
	//     to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
	//     -f ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz -head HEAD`.
	if !skipBuild {
		if err := os.Chdir(srcDir); err != nil {
			return false, err
		}
		head := shDNS.HeadBuf()
		distFile := filepath.Join("..", "dists", fn)
		if pkg.Key != "" {
			key, err := pkg.GetKey()
			if err != nil {
				return false, err
			}
			err = archive.ApplyEncryptedFile(def.HashchainFile, def.PatchDir,
				distFile, &head, key)
			if err != nil {
				return false, err
			}
		} else {
			err = archive.ApplyFile(def.HashchainFile, def.PatchDir, distFile, &head)
			if err != nil {
				return false, err
			}
		}
		c, err := hashchain.ReadFile(def.HashchainFile)
		if err != nil {
			return false, err
		}
		if err := c.Close(); err != nil {
			return false, err
		}
		if err := c.Apply(&head); err != nil {
			return false, err
		}
	}

	// 11. If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
	//     contains any .secpkg files, ensure these secure dependencies are
	//     installed and up-to-date. If at least one dependency was updated, set
	//     SKIP_BUILD to false.
	depUpdated, err := ensure(ctx, visited, name)
	if err != nil {
		return false, err
	}
	if depUpdated {
		skipBuild = false
	}

	// 12. If not SKIP_BUILD, call `make prefix=~/.config/secpkg/local uninstall` in
	//     ~/.config/secpkg/pkgs/NAME/installed
	installedDir := filepath.Join(pkgDir, "installed")
	localDir := filepath.Join(homedir.SecPkg(), "local")
	if !skipBuild {
		if err := os.Chdir(installedDir); err != nil {
			return false, err
		}
		if err := gnumake.Uninstall(localDir); err != nil {
			return false, err
		}
	}

	// 13. If not SKIP_BUILD, `rm -rf ~/.config/secpkg/pkgs/NAME/build`
	buildDir := filepath.Join(pkgDir, "build")
	if !skipBuild {
		if err := os.RemoveAll(buildDir); err != nil {
			return false, err
		}
	}

	// 14. If not SKIP_BUILD,
	//     `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`
	if !skipBuild {
		if err := file.CopyDir(srcDir, buildDir); err != nil {
			return false, err
		}
	}

	// 16. If not SKIP_BUILD, call `make prefix=~/.config/secpkg/local` in
	//     ~/.config/secpkg/pkgs/NAME/build
	if !skipBuild {
		if err := os.Chdir(buildDir); err != nil {
			os.RemoveAll(pkgDir)
			return false, err
		}
		if err := gnumake.Call(localDir); err != nil {
			return false, err
		}
	}

	// 16. If not SKIP_BUILD, call `make prefix= ~/.config/secpkg/local install` in
	//     ~/.config/secpkg/pkgs/NAME/build
	if !skipBuild {
		if err := gnumake.Install(localDir); err != nil {
			return false, err
		}
	}

	// 17. If not SKIP_BUILD,
	//     `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`
	if !skipBuild {
		if err := os.RemoveAll(installedDir); err != nil {
			return false, err
		}
		if err := os.Rename(buildDir, installedDir); err != nil {
			return false, err
		}
	}

	// 18. Update signed head:
	//
	//      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
	//               ~/.config/secpkg/pkgs/NAME/previous_signed_head`
	//      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
	if err := shDNS.RotateFile(pkgDir); err != nil {
		return false, nil
	}

	// 19. The software has been successfully updated.
	if skipBuild {
		fmt.Printf("package '%s' already up-to-date\n", name)
		return false, nil
	}
	return true, nil
}

// Update package with name, see specification for details.
func Update(ctx context.Context, name string) error {
	visited := make(map[string]bool)
	visited[name] = true
	_, err := update(ctx, visited, name)
	return err
}
