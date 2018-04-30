// Package tree implements functions to hash directory trees.
package tree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// EmptyHash is the hash an empty directory tree (in hex notation).
const EmptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// List returns a list in lexical order of newline separated hashes of all
// files in the file tree rooted at root, except for the paths in
// excludePaths. Directories are only implicitly listed (i.e., if they
// contain files). Entries start with 'f' if it is a regular file (read and
// write permission for user) and with 'x' if it is an executable (read,
// write, and executabele for user).
//
// The deterministic list serves as the basis for a hash of a directory tree.
//
// The directory tree can only contain directories or regular files.
//
// Example list:
//	f 7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730 bar/baz.txt
//	x b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c foo.txt
func List(root string, excludePaths []string) ([]byte, error) {
	var b bytes.Buffer
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
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		fmt.Fprintf(&b, "%c %x %s\n", m, h.Sum(nil), canonical)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Hash returns a SHA256 hash of all files and directories in the file tree
// rooted at root, except for the paths in excludePaths. The result of the
// List function serves as a deterministic input if the hash function.
func Hash(root string, excludePaths []string) ([]byte, error) {
	l, err := List(root, excludePaths)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(l)
	return h[:], nil
}
