// ccfindhead finds a given head in a hash chain file.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

func findHead(hashchainFile, headInHex string) error {
	head, err := hex.DecodeString(headInHex)
	if err != nil {
		return err
	}
	if len(head) != 32 {
		return errors.New("head_in_hex is not 32 bytes long")
	}
	f, err := os.Open(hashchainFile)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineHash := sha256.Sum256(scanner.Bytes())
		if bytes.Equal(lineHash[:], head) {
			fmt.Println("head found")
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return errors.New("cannot find head")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s hashchain_file head_in_hex\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	if err := findHead(os.Args[1], os.Args[2]); err != nil {
		fatal(err)
	}
}
