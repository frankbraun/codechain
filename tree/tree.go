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
// files and directories in the file tree rooted at root, except for the paths
// in excludePaths. This serves as the basis for a hash of a directory tree.
//
// The tree rooted at root can only contain directories or regular files.
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
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				return err
			}
			fmt.Fprintf(&b, "f %3o %x", perm, h.Sum(nil))
		} else {
			fmt.Fprintf(&b, "d %3o", perm)
		}
		fmt.Fprintf(&b, " %s\n", canonical)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Hash returns a deterministic hash of all files and directories in the file tree rooted at root, except for the paths
// in excludePaths (using the List function).
func Hash(root string, excludePaths []string) ([]byte, error) {
	l, err := List(root, excludePaths)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(l)
	return h[:], nil
}
