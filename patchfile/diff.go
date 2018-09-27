// Package patchfile implements a robust patchfile format for directory trees.
package patchfile

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
)

func writeDeleteFile(w io.Writer, entry tree.ListEntry) {
	fmt.Fprintf(w, "- %c %x %s\n", entry.Mode, entry.Hash, entry.Filename)
}

func writeAddFile(w io.Writer, dir string, entry tree.ListEntry) error {
	fmt.Fprintf(w, "+ %c %x %s\n", entry.Mode, entry.Hash, entry.Filename)
	filename := filepath.Join(dir, entry.Filename)
	isBinary, err := file.IsBinary(filename)
	if err != nil {
		return err
	}
	if isBinary {
		err := ascii85Diff(w, filename)
		if err != nil {
			return err
		}
	} else {
		err := dmpDiff(w, "", filename)
		if err != nil {
			return err
		}

	}
	return nil
}

// Diff computes a patch w between the directory trees rooted at a and b.
// If a and b have the same tree hash ErrNoDifference is returned.
// In case of error, some data might have been written to w already.
func Diff(w io.Writer, a, b string, excludePaths []string) error {
	listA, err := tree.List(a, excludePaths)
	if err != nil {
		return err
	}
	listB, err := tree.List(b, excludePaths)
	if err != nil {
		return err
	}
	hashA := tree.HashList(listA)
	hashB := tree.HashList(listB)
	if bytes.Equal(hashA[:], hashB[:]) {
		return ErrNoDifference
	}
	fmt.Fprintf(w, "codechain patchfile version %d\n", Version)
	fmt.Fprintf(w, "treehash %x\n", hashA[:])
	idxA := 0
	idxB := 0
	for idxA < len(listA) && idxB < len(listB) {
		entryA := listA[idxA]
		entryB := listB[idxB]
		if entryA.Filename == entryB.Filename {
			if !bytes.Equal(entryA.Hash[:], entryB.Hash[:]) ||
				entryA.Mode != entryB.Mode {
				fmt.Fprintf(w, "- %c %x %s\n", entryA.Mode, entryA.Hash, entryA.Filename)
				fmt.Fprintf(w, "+ %c %x %s\n", entryB.Mode, entryB.Hash, entryB.Filename)
			}
			if !bytes.Equal(entryA.Hash[:], entryB.Hash[:]) {
				filenameA := filepath.Join(a, entryA.Filename)
				filenameB := filepath.Join(b, entryB.Filename)
				isBinaryA, err := file.IsBinary(filenameA)
				if err != nil {
					return err
				}
				isBinaryB, err := file.IsBinary(filenameB)
				if err != nil {
					return err
				}
				if isBinaryA || isBinaryB {
					err := ascii85Diff(w, filenameB)
					if err != nil {
						return err
					}
				} else {
					err := dmpDiff(w, filenameA, filenameB)
					if err != nil {
						return err
					}

				}
			}
		} else if entryA.Filename < entryB.Filename {
			writeDeleteFile(w, entryA)
			idxA++
			continue
		} else { // entryA.Filename > entryB.Filename
			if err := writeAddFile(w, b, entryB); err != nil {
				return err
			}
			idxB++
			continue
		}
		idxA++
		idxB++
	}
	for idxA < len(listA) {
		writeDeleteFile(w, listA[idxA])
		idxA++
	}
	for idxB < len(listB) {
		if err := writeAddFile(w, b, listB[idxB]); err != nil {
			return err
		}
		idxB++
	}
	fmt.Fprintf(w, "treehash %x\n", hashB[:])
	return nil
}
