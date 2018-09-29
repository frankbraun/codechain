package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
)

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func install(pkg *secpkg.Package) error {
	txts, err := net.LookupTXT("_codechain." + pkg.DNS)
	if err != nil {
		return err
	}
	// Parse TXT records and look for signed head
	var sh *ssot.SignedHead
	for _, txt := range txts {
		sh, err = ssot.Unmarshal(txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot unmarshal: %s\n", txt)
			continue
		}
		fmt.Println(sh.Head())
		break /// TXT record found
	}
	if sh == nil {
		return errors.New("secpkg: no valid TXT record found")
	}
	// TODO: trust pubkey on first use
	// TODO: compare pubkey with stored one
	// TODO: pubkey rotation

	// download distribution
	filename := sh.Head() + ".tar.gz"
	url := pkg.URL + "/" + filename
	fmt.Printf("download %s\n", url)
	if err := downloadFile(filename, url); err != nil {
		return err
	}
	return nil
}

// Install implements the secpkg 'install' command.
func Install(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s project.secpkg\n", argv0)
		fmt.Fprintf(os.Stderr, "Download, verify, and install package defined by project.secpkg.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	pkg, err := secpkg.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Println(pkg.Marshal())
	return install(pkg)
}
