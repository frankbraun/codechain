// secpkg installs and updates secure Codechain packages.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/secpkg/command"
)

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage: %s install project.secpkg\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s update package_name\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s uninstall package_name\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s list\n", cmd)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	argv0 := os.Args[0] + " " + os.Args[1]
	args := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "install":
		err = command.Install(argv0, args...)
	case "update":
		err = command.Update(argv0, args...)
	case "uninstall":
		err = command.Uninstall(argv0, args...)
	case "list":
		err = command.List(argv0, args...)
	default:
		usage()
	}
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
			os.Exit(1)
		}
		os.Exit(2)
	}
}
