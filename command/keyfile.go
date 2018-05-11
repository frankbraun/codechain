package command

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/terminal"
)

func changePassphrase(filename string, sec, sig *[64]byte, comment []byte) error {
	pass, err := terminal.ReadPassphrase(syscall.Stdin, true)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	tmpfile := filename + ".new"
	os.Remove(tmpfile) // ignore error
	// create new keyfile
	if err := keyfile.Create(tmpfile, pass, *sec, *sig, comment); err != nil {
		return err
	}
	// move temp. file in place
	return os.Rename(tmpfile, filename)
}

func listKeys() error {
	homeDir := filepath.Join(codechainHomeDir(), secretsDir)
	files, err := ioutil.ReadDir(homeDir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		filename := filepath.Join(homeDir, fi.Name())
		fmt.Println(filename)
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		line, err := bufio.NewReader(f).ReadString('\n')
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
		fields := strings.SplitN(line, " ", 3)
		if len(fields) == 3 {
			fmt.Print(fields[2])
		}
	}
	return nil
}

// KeyFile implements the 'keyfile' command.
func KeyFile(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Show pubkey, signature, and comment for encrypted secret key file.\n")
		fs.PrintDefaults()
	}
	change := fs.Bool("c", false, "Change passphrase")
	list := fs.Bool("l", false, "List keyfiles")
	seckey := fs.String("s", "", "Secret key file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *change && *list {
		return fmt.Errorf("%s: options -c and -l exclude each other", argv0)
	}
	if *seckey != "" && *list {
		return fmt.Errorf("%s: options -s and -l exclude each other", argv0)
	}
	if *seckey == "" && !*list {
		return fmt.Errorf("%s: option -s is mandatory", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if *list {
		return listKeys()
	}
	sec, sig, comment, err := seckeyRead(*seckey)
	if err != nil {
		return err
	}
	if *change {
		fmt.Printf("%s read, please provide new ", *seckey)
		if err := changePassphrase(*seckey, sec, sig, comment); err != nil {
			return err
		}
		fmt.Println("passphrase changed")
	} else {
		fmt.Println("public key with signature and optional comment:")
		fmt.Printf("%s %s", base64.Encode(sec[32:]), base64.Encode(sig[:]))
		if len(comment) > 0 {
			fmt.Printf(" '%s'", string(comment))
		}
		fmt.Println("")
	}
	return nil
}
