package secpkg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

// Install pkg, see specification for details.
func (pkg *Package) Install() error {
	// 1. has already been done by calling Load()

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
	return file.Download(filename, url)
}
