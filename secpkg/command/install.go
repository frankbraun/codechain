package command

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/frankbraun/codechain/secpkg"
)

func install(pkg *secpkg.Package) error {
	// TODO: txts, err := net.LookupTXT("_codechain." + pkg.DNS)
	txts, err := net.LookupTXT("_test.frankbraun.org")
	if err != nil {
		return err
	}
	for _, txt := range txts {
		fmt.Println(txt)
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
