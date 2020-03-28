package secpkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/gnumake"
	"github.com/frankbraun/codechain/util/hex"
)

type dnsRecord struct {
	DNS string
	sh  ssot.SignedHead
}

type dnsRecords []dnsRecord

func (d dnsRecords) Len() int           { return len(d) }
func (d dnsRecords) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d dnsRecords) Less(i, j int) bool { return d[i].sh.Line() > d[j].sh.Line() }

func (pkg *Package) install(
	ctx context.Context,
	res Resolver,
	homedir string,
	visited map[string]bool,
) error {
	// 1. Has already been done by calling Load().

	// 2. Make sure the project has not been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME does not exist.
	pkgDir := filepath.Join(homedir, "pkgs", pkg.Name)
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

	// 5. Get next DNS entry from DNS_RECORDS.
	var dnsRecords dnsRecords
	for _, DNS := range pkg.DNSRecords() {
		dnsRecords = append(dnsRecords, dnsRecord{DNS: DNS})
	}
	for i, dnsRecord := range dnsRecords {
		// 6. Query TXT record from _codechain-head.DNS and validate the signed head
		//    contained in it (see ssot package).
		sh, err := res.LookupHead(ctx, dnsRecord.DNS)
		if err != nil {
			os.RemoveAll(pkgDir)
			return err
		}
		dnsRecords[i].sh = sh

		// 7. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head.DNS
		signedHead := filepath.Join(pkgDir, ssot.File+"."+dnsRecord.DNS)
		err = ioutil.WriteFile(signedHead, []byte(sh.Marshal()+"\n"), 0644)
		if err != nil {
			os.RemoveAll(pkgDir)
			return err
		}
		fmt.Printf("%s: written\n", signedHead)
	}

	// 8. Sort DNS_RECORDS in descending order according to the last signed line
	//    number (signed head version 2 or higher).
	sort.Sort(dnsRecords)

	// 9. Get next DNS entry from DNS_RECORDS. If no such entry exists, exit with
	//    error.
	i := 0
	var dnsRecord dnsRecord
_9:
	if i < len(dnsRecords) {
		dnsRecord = dnsRecords[i]
		fmt.Printf("get DNS: %s\n", dnsRecord.DNS)
		i++
	} else {
		os.RemoveAll(pkgDir)
		return fmt.Errorf("no valid DNS entry found")
	}

	// 10. Query all TXT records from _codechain-url.DNS and save it as URLs.
	//     If no such record exists: Goto 9.
	URLs, err := res.LookupURLs(ctx, dnsRecord.DNS)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		goto _9
	}

	// 11. Select next URL from URLs. If no such URL exists: Goto 9.
	j := 0
	var URL string
_11:
	if j < len(URLs) {
		URL = URLs[j]
		fmt.Printf("try URL: %s\n", URL)
		j++
	} else {
		goto _9
	}

	// 12. Download distribution file from URL/HEAD_SSOT.tar.gz and save it to
	//     ~/.config/secpkg/pkgs/NAME/dists
	//     If it fails: Goto 11.
	distDir := filepath.Join(pkgDir, "dists")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	var encSuffix string
	if pkg.Key != "" {
		encSuffix = ".enc"
	}
	fn = dnsRecord.sh.Head() + ".tar.gz" + encSuffix
	filename := filepath.Join(distDir, fn)
	url := URL + "/" + fn
	fmt.Printf("download %s\n", url)
	err = res.Download(filename, url)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		goto _11
	}

	// 13. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz
	//     to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
	//     -f ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz -head HEAD_SSOT`
	//     If it fails: Goto 11.
	srcDir := filepath.Join(pkgDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := os.Chdir(srcDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	head := dnsRecord.sh.HeadBuf()
	distFile := filepath.Join("..", "dists", fn)
	if pkg.Key != "" {
		key, err := pkg.GetKey()
		if err != nil {
			return err
		}
		err = archive.ApplyEncryptedFile(def.UnoverwriteableHashchainFile, def.PatchDir,
			distFile, &head, key)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			goto _11
		}
	} else {
		err = archive.ApplyFile(def.UnoverwriteableHashchainFile, def.PatchDir, distFile, &head)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			goto _11
		}
	}
	c, err := hashchain.ReadFile(def.UnoverwriteableHashchainFile)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := c.Close(); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	if err := c.Apply(&head, def.PatchDir); err != nil {
		fmt.Printf("error: %s\n", err)
		goto _11
	}

	// 14. Make sure HEAD_PKG is contained in
	//     ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain
	//     If it fails: Goto 11.
	h, err := hex.Decode(pkg.Head, 32)
	if err != nil {
		os.RemoveAll(pkgDir)
		return err
	}
	copy(head[:], h)
	if err := c.CheckHead(head); err != nil {
		fmt.Printf("error: %s\n", err)
		goto _11
	}

	// 15. If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
	//     contains any .secpkg files, ensure these secure dependencies are
	//     installed and up-to-date.
	if _, err := ensure(ctx, res, homedir, visited, pkg.Name); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 16. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`
	buildDir := filepath.Join(pkgDir, "build")
	if err := file.CopyDir(srcDir, buildDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 17. Call `make prefix=~/.config/secpkg/local` in
	//     ~/.config/secpkg/pkgs/NAME/build
	localDir := filepath.Join(homedir, "local")
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

	// 18. Call `make prefix=~/.config/secpkg/local install` in
	//     ~/.config/secpkg/pkgs/NAME/build
	if err := gnumake.Install(localDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 19. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`
	installedDir := filepath.Join(pkgDir, "installed")
	if err := os.Rename(buildDir, installedDir); err != nil {
		os.RemoveAll(pkgDir)
		return err
	}

	// 20. If the file ~/.config/secpkg/pkgs/NAME/installed/.secpkg exists,
	//     `cp -f ~/.config/secpkg/pkgs/NAME/installed/.secpkg
	//            ~/.config/secpkg/pkgs/NAME/.secpkg`
	insSecpkgFile := filepath.Join(installedDir, File)
	exists, err = file.Exists(insSecpkgFile)
	if err != nil {
		return err
	}
	if exists {
		defSecpkgFile := filepath.Join(pkgDir, File)
		newSecpkgFile := filepath.Join(pkgDir, File+".new")
		if err := os.RemoveAll(newSecpkgFile); err != nil {
			return err
		}
		if err := file.Copy(insSecpkgFile, newSecpkgFile); err != nil {
			return err
		}
		if err := os.Rename(newSecpkgFile, defSecpkgFile); err != nil {
			return err
		}
	}

	return nil
}

// Install pkg, see specification for details.
func (pkg *Package) Install(
	ctx context.Context,
	res Resolver,
	homedir string,
) error {
	visited := make(map[string]bool)
	visited[pkg.Name] = true
	return pkg.install(ctx, res, homedir, visited)
}
