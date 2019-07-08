/*
cctreehash calculates and prints the tree hash of the current directory in hex.

This file is frozen, do not change it.

This stand alone tool helps bootstrapping Codechain,
see https://github.com/frankbraun/codechain/blob/master/doc/bootstrapping.md.

To calculate the hash of a directory tree (a tree hash) a list of all files in
the directory root (a tree list) is created as follows.

All the files below the root of the directory tree are traversed in lexical
order (with filepath.Walk) and printed in this format:

  m xxx filename

Where:
  m        is the mode ('f' or 'x')
  xxx      is the SHA256 hash for the file in hex notation
  filename is the file name with directory prefix starting at root

Example list:
  f 7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730 bar/baz.txt
  x b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c foo.txt

The fields are separated with single white space characters and the lines are
separated with single newline characters.

Directories are only implicitly listed (i.e., if they
contain files). Entries start with 'f' if it is a regular file (read and
write permission for user) and with 'x' if it is an executable (read,
write, and executable for user).

The directory tree must only contain directories, regular files, or executables.

The deterministic tree list serves as the basis for a hash of a directory tree
(the tree hash), which is the SHA256 hash of the tree list in hex notation.
*/
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExcludePaths is the default list of paths not considered by Codechain.
// Do not ever change this list! It will break existing Codechains.
var ExcludePaths = []string{
	".codechain",
	".git",
	".gitignore",
	".travis.yml",
}

// SHA256 returns the SHA256 hash of the file with given path.
func SHA256(path string) (*[32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	sum := h.Sum(nil)
	var hash [32]byte
	copy(hash[:], sum)
	return &hash, nil
}

// ListEntry describes a directory tree entry.
type ListEntry struct {
	Mode     rune     // 'f' (regular) or 'x' (binary)
	Filename string   // Including directory path starting from root
	Hash     [32]byte // SHA256 hash
}

// List returns a list in lexical order of ListEntry structs of all files in
// the file tree rooted at root. See the ListBytes function for details.
//
// This is a convenience function to make a tree list accessible without
// having to parse tree list entries.
func List(root string, excludePaths []string) ([]ListEntry, error) {
	var entries []ListEntry
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("%s: neither directory nor normal file", path)
		}
		if path == root {
			return nil
		}
		canonical := path
		if root != "." {
			canonical = strings.TrimPrefix(path, root)
			canonical = strings.TrimPrefix(canonical, string(filepath.Separator))
		}
		canonical = filepath.ToSlash(canonical)
		if excludePaths != nil {
			for _, excludePath := range excludePaths {
				if excludePath == canonical {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}
		perm := info.Mode().Perm() & os.ModePerm
		if info.IsDir() {
			if perm&0700 != 0700 {
				return fmt.Errorf("%s: directory doesn't have all user permissions", path)
			}
			return nil
		}
		var m rune
		if perm&0100 == 0100 {
			if perm&0700 != 0700 {
				return fmt.Errorf("%s: executable is not readable and writable", path)
			}
			m = 'x' // executable
		} else {
			if perm&0010 > 0 {
				return fmt.Errorf("%s: regular file is executable for group, but not for user", path)
			}
			if perm&0001 > 0 {
				return fmt.Errorf("%s: regular file is executable for other, but not for user", path)
			}
			if perm&0600 != 0600 {
				return fmt.Errorf("%s: regular file is not readable and writable", path)
			}
			m = 'f' // regular file
		}
		h, err := SHA256(path)
		if err != nil {
			return err
		}
		entries = append(entries, ListEntry{m, canonical, *h})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// PrintList prints a list of (sorted) entries in the canonical tree list
// format.
//
// This is a convience function to bring a list of entries returned by the
// List functions into the format returned by the ListBytes function.
func PrintList(entries []ListEntry) []byte {
	var b bytes.Buffer
	for _, e := range entries {
		fmt.Fprintf(&b, "%c %x %s\n", e.Mode, e.Hash[:], e.Filename)
	}
	return b.Bytes()
}

// ListBytes returns a list in lexical order of newline separated hashes of
// all files in the file tree rooted at root in the canonical format, except
// for the paths in excludePaths (the tree list).
func ListBytes(root string, excludePaths []string) ([]byte, error) {
	entries, err := List(root, excludePaths)
	if err != nil {
		return nil, err
	}
	return PrintList(entries), nil
}

// Hash returns a SHA256 hash of all files and directories in the file tree
// rooted at root, except for the paths in excludePaths (the tree hash).
func Hash(root string, excludePaths []string) (*[32]byte, error) {
	l, err := ListBytes(root, excludePaths)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(l)
	return &h, nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func main() {
	// calculate tree hash
	hash, err := Hash(".", ExcludePaths)
	if err != nil {
		fatal(err)
	}
	// print tree hash in hex
	fmt.Printf("%x\n", hash[:])
}
