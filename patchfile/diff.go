package patchfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
)

// writeFileDeletion writes the tree list entry as a file deletion to w.
func writeFileDeletion(w io.Writer, entry tree.ListEntry) {
	// file deletion
	fmt.Fprintf(w, "- %c %x %s\n", entry.Mode, entry.Hash, entry.Filename)
}

// writeFileAddition writes the tree list entry (in root dir) as a file
// addition to w.
//
// It determines if the file in entry is binary or UTF-8 and encodes it
// accordingly as an "ascii85" or "dmppatch" patch.
func writeFileAddition(version int, w io.Writer, dir string, entry tree.ListEntry) error {
	// file addition
	fmt.Fprintf(w, "+ %c %x %s\n", entry.Mode, entry.Hash, entry.Filename)
	// filename regarding the root dir
	filename := filepath.Join(dir, entry.Filename)
	// check if the file is binary
	isBinary, err := file.IsBinary(filename)
	if err != nil {
		return err
	}
	if isBinary {
		// write "ascii85" encoding
		err := ascii85Diff(w, filename)
		if err != nil {
			return err
		}
	} else if version > 1 {
		// write "utf8file" patch
		err := utf8fileDiff(w, filename)
		if err != nil {
			return err
		}
	} else {
		// write "dmppatch" patch
		clean, err := dmpDiff(w, "", filename)
		if err != nil {
			return err
		}
		if !clean {
			return ErrDiffNotClean
		}
	}
	return nil
}

// writeFileDiff writes the diff between the tree list entryA (in directory a)
// and tree list entryB (in directory b) as a file diff to w.
//
// If neither the file hash nor the file mode of entryA and entryB differ, the
// functions returns nil without writing anything to w.
//
// If the file hashes differ, the function determines if either of the files
// is binary or both are UTF-8 and encodes the diff accordingly as an
// "ascii85" or "dmppatch" patch.
func writeFileDiff(version int, w io.Writer, a, b string, entryA, entryB tree.ListEntry) error {
	// Assert that file diffs are only used if the file names are the same.
	if !(entryA.Filename == entryB.Filename) {
		panic(errors.New("patchfile: entryA.Filename != entryB.Filename"))
	}
	// Write a file diff, if the file hash or the file mode changed.
	if !bytes.Equal(entryA.Hash[:], entryB.Hash[:]) ||
		entryA.Mode != entryB.Mode {
		fmt.Fprintf(w, "- %c %x %s\n", entryA.Mode, entryA.Hash, entryA.Filename)
		fmt.Fprintf(w, "+ %c %x %s\n", entryB.Mode, entryB.Hash, entryB.Filename)
	}
	// Write actual patch, if the file hash changed.
	if !bytes.Equal(entryA.Hash[:], entryB.Hash[:]) {
		// Check if either of the files is binary.
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
			// write "ascii85" encoding
			err := ascii85Diff(w, filenameB)
			if err != nil {
				return err
			}
		} else {
			// write "dmppatch" patch, if possible
			clean, err := dmpDiff(w, filenameA, filenameB)
			if err != nil {
				return err
			}
			if !clean {
				// for version 1 we have to give up here.
				if version == 1 {
					return ErrDiffNotClean
				}
				// write "utf8file" patch instead
				if err := utf8fileDiff(w, filenameB); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Diff computes a patch between the directory trees rooted at a and b and
// writes it to w. If a and b have the same tree hash ErrNoDifference is
// returned. In case of error, some data might have been written to w already.
// The paths given in excludePaths are excluded from all tree hash calculations.
func Diff(version int, w io.Writer, a, b string, excludePaths []string) error {
	// only support version 1 and 2
	if version != 1 && version != 2 {
		return ErrHeaderVersion
	}
	// Calculate tree list of "source" directory tree.
	listA, err := tree.List(a, excludePaths)
	if err != nil {
		return err
	}
	// Calculate tree list of "target" directory tree.
	listB, err := tree.List(b, excludePaths)
	if err != nil {
		return err
	}
	// Hash directories trees and compare them.
	hashA := tree.HashList(listA)
	hashB := tree.HashList(listB)
	if bytes.Equal(hashA[:], hashB[:]) {
		return ErrNoDifference
	}
	// version line
	fmt.Fprintf(w, "codechain patchfile version %d\n", version)
	// initial treehash line
	fmt.Fprintf(w, "treehash %x\n", hashA[:])
	idxA := 0
	idxB := 0
	for idxA < len(listA) && idxB < len(listB) {
		entryA := listA[idxA]
		entryB := listB[idxB]
		if entryA.Filename == entryB.Filename {
			// write file diff, if necessary
			if err := writeFileDiff(version, w, a, b, entryA, entryB); err != nil {
				return err
			}
		} else if entryA.Filename < entryB.Filename {
			writeFileDeletion(w, entryA)
			idxA++
			continue
		} else { // entryA.Filename > entryB.Filename
			if err := writeFileAddition(version, w, b, entryB); err != nil {
				return err
			}
			idxB++
			continue
		}
		idxA++
		idxB++
	}
	for idxA < len(listA) {
		writeFileDeletion(w, listA[idxA])
		idxA++
	}
	for idxB < len(listB) {
		if err := writeFileAddition(version, w, b, listB[idxB]); err != nil {
			return err
		}
		idxB++
	}
	// final treehash line
	fmt.Fprintf(w, "treehash %x\n", hashB[:])
	return nil
}
