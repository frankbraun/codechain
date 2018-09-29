package command

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

func install(pkg *secpkg.Package) error {
	txts, err := net.LookupTXT(def.CodechainTXTName + pkg.DNS)
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
	if err := file.Download(filename, url); err != nil {
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
