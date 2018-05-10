// Package patchfile implements a robust patchfile format for directory trees.
package patchfile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
)

// Version is the maximum codechain patchfile version this package can parse.
const Version = 1

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

type state int

const (
	start state = iota + 1
	treehash
	fileDiff
	secondFileDiff
	addFile
	diffFile
	result
	terminal
)

// codechain patchfile version 1
func procStart(line string) (state, error) {
	fields := strings.SplitN(line, " ", 4)
	if len(fields) != 4 {
		return 0, ErrHeaderFieldsNum
	}
	if fields[0] != "codechain" || fields[1] != "patchfile" || fields[2] != "version" {
		return 0, ErrHeaderFieldsText
	}
	ver, err := strconv.Atoi(fields[3])
	if err != nil {
		return 0, err
	}
	if ver != 1 { // only support Version = 1
		return 0, ErrHeaderVersion
	}
	return treehash, nil
}

// treehash hex_hash
func procTreeHash(line, dir string, excludePaths []string) (state, error) {
	fields := strings.SplitN(line, " ", 2)
	if len(fields) != 2 {
		return 0, ErrTreeHashFieldsNum
	}
	if fields[0] != "treehash" {
		return 0, ErrTreeHashFieldsText
	}
	h, err := hex.Decode(fields[1], 32)
	if err != nil {
		return 0, err
	}
	treeHash, err := tree.Hash(dir, excludePaths)
	if err != nil {
		return 0, err
	}
	if !bytes.Equal(treeHash[:], h) {
		return 0, ErrTreeHashStartMismatch
	}
	return fileDiff, nil
}

type mode int

const (
	regularFile mode = iota + 1
	binaryFile
)

// + f hex_hash filename # add file, filename must not exist
//
// or
//
// - f hex_hash filename # delete file
//
// or
//
// - f hex_hash filename_a # possible mode change
// + x hex_hash filename_b # if filename_a and filname_b differ, filename_b must not exist
func procFileDiff(line string, dir string, prevDiffInfo *diffInfo) (state, *diffInfo, error) {
	fields := strings.SplitN(line, " ", 4)
	if len(fields) != 4 {
		return 0, nil, ErrFileFieldsNum
	}
	if fields[0] != "-" && fields[0] != "+" {
		return 0, nil, ErrFileField0
	}
	if fields[1] != "f" && fields[1] != "x" {
		return 0, nil, ErrFileField1
	}
	hash, err := hex.Decode(fields[2], 32)
	if err != nil {
		return 0, nil, err
	}
	name := fields[3]
	var mode mode
	if fields[1] == "f" {
		mode = regularFile
	} else {
		mode = binaryFile
	}
	if prevDiffInfo != nil {
		if name != prevDiffInfo.name {
			// move, make sure target doesn't exist
			fn := filepath.Join(dir, name)
			exists, err := file.Exists(fn)
			if err != nil {
				return 0, nil, err
			}
			if exists {
				return 0, nil, ErrMoveTargetFileExists
			}
		}
		if bytes.Equal(hash[:], prevDiffInfo.hash) {
			if mode != prevDiffInfo.mode {
				// chmod
				var perm os.FileMode
				if mode == regularFile {
					perm = 0644
				} else {
					perm = 0755
				}
				if err := os.Chmod(filepath.Join(dir, name), perm); err != nil {
					return 0, nil, err
				}
			}
			if name != prevDiffInfo.name {
				// delete and then add
				oldpath := filepath.Join(dir, prevDiffInfo.name)
				if err := os.Remove(oldpath); err != nil {
					return 0, nil, err
				}
				return addFile, &diffInfo{mode, hash, name}, nil
			}
			return fileDiff, nil, nil
		}
		// diff
		return diffFile, &diffInfo{mode, hash, name}, nil
	} else if fields[0] == "+" {
		// add file
		fn := filepath.Join(dir, name)
		exists, err := file.Exists(fn)
		if err != nil {
			return 0, nil, err
		}
		if exists {
			return 0, nil, ErrAddTargetFileExists
		}
		return addFile, &diffInfo{mode, hash, name}, nil
	}
	// else: delete or diff?
	return secondFileDiff, &diffInfo{mode, hash, name}, nil
}

func procResult(line, dir string, excludePaths []string) (state, error) {
	if _, err := procTreeHash(line, dir, excludePaths); err != nil {
		if err == ErrTreeHashStartMismatch {
			return 0, ErrTreeHashFinishMismatch
		}
		return 0, err
	}
	return terminal, nil
}

type applyFunc func(w io.Writer, text string, patch []byte) error

func apply(dir string, buf []byte, state state, prev, cur *diffInfo, applyFunc applyFunc) error {
	var text string
	fileB := filepath.Join(dir, cur.name)
	if state == addFile {
		if err := os.MkdirAll(filepath.Dir(fileB), 0755); err != nil {
			return err
		}
	} else {
		fileA := filepath.Join(dir, prev.name)
		hash, err := tree.SHA256(fileA)
		if err != nil {
			return err
		}
		if !bytes.Equal(hash[:], prev.hash) {
			return ErrFileHashMismatchBefore
		}
		buf, err := ioutil.ReadFile(fileA)
		if err != nil {
			return err
		}
		text = string(buf)
	}
	var flag int
	if state == addFile {
		flag = os.O_CREATE | os.O_EXCL | os.O_WRONLY
	} else {
		flag = os.O_TRUNC | os.O_WRONLY
	}
	var perm os.FileMode
	if cur.mode == regularFile {
		perm = 0644
	} else {
		perm = 0755
	}
	f, err := os.OpenFile(fileB, flag, perm)
	if err != nil {
		return err
	}
	if err := applyFunc(f, text, buf); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	hash, err := tree.SHA256(fileB)
	if err != nil {
		return err
	}
	if !bytes.Equal(hash[:], cur.hash) {
		return ErrFileHashMismatchAfter
	}
	return nil
}

type diffInfo struct {
	mode
	hash []byte
	name string
}

// Apply applies the given patch r to the directory tree dir.
func Apply(dir string, r io.Reader, excludePaths []string) error {
	var (
		prevDiffInfo *diffInfo
		curDiffInfo  *diffInfo
		err          error
	)
	s := bufio.NewScanner(r)
	buf := make([]byte, bufio.MaxScanTokenSize)
	s.Buffer(buf, 64*1024*1024) // 64MB, entire files can be encoded as single lines
	state := start
	for s.Scan() {
		line := s.Text()
		//fmt.Println(line)
		switch state {
		case start:
			state, err = procStart(line)
			if err != nil {
				return err
			}
		case treehash:
			state, err = procTreeHash(line, dir, excludePaths)
			if err != nil {
				return err
			}
		case fileDiff:
			fields := strings.SplitN(line, " ", 2)
			lookAhead := fields[0]
			if lookAhead == "treehash" {
				state, err = procResult(line, dir, excludePaths)
				if err != nil {
					return err
				}
			} else {
				prevDiffInfo = nil
				state, curDiffInfo, err = procFileDiff(line, dir, prevDiffInfo)
				if err != nil {
					return err
				}
			}
		case secondFileDiff:
			fields := strings.SplitN(line, " ", 2)
			lookAhead := fields[0]
			if lookAhead == "-" || lookAhead == "treehash" {
				// delete
				fn := filepath.Join(dir, curDiffInfo.name)
				hash, err := tree.SHA256(fn)
				if err != nil {
					return err
				}
				if !bytes.Equal(hash[:], curDiffInfo.hash) {
					return fmt.Errorf("patchfile: hash of file '%s' to delete doesn't match",
						prevDiffInfo.name)
				}
				if err := os.Remove(fn); err != nil {
					return err
				}
				// reset
				curDiffInfo = nil
			}
			if lookAhead == "treehash" {
				state, err = procResult(line, dir, excludePaths)
				if err != nil {
					return err
				}
			} else {
				prevDiffInfo = curDiffInfo
				state, curDiffInfo, err = procFileDiff(line, dir, prevDiffInfo)
				if err != nil {
					return err
				}
			}
		case addFile:
			fallthrough
		case diffFile:
			fields := strings.SplitN(line, " ", 2)
			lookAhead := fields[0]
			numLines, err := strconv.Atoi(fields[1])
			if err != nil {
				return ErrDiffLinesParse
			}
			switch lookAhead {
			case "ascii85":
				if numLines < 1 {
					return ErrDiffLinesNonPositive
				}
			case "dmppatch":
				if numLines < 0 {
					return ErrDiffLinesNegative
				}
			default:
				return ErrDiffModeUnknown
			}
			// read lines
			var lines []string
			for i := 0; i < numLines; i++ {
				if s.Scan() {
					line := s.Text()
					//fmt.Println(line)
					lines = append(lines, line)
				} else {
					// check if we have a scanner error first
					if err := s.Err(); err != nil {
						return err
					}
					return ErrPrematureDiffEnd
				}
			}
			switch lookAhead {
			case "ascii85":
				buf := strings.Join(lines, "")
				err = apply(dir, []byte(buf), state, prevDiffInfo, curDiffInfo, ascii85Apply)
				if err != nil {
					return err
				}
				// reset
				prevDiffInfo = nil
				curDiffInfo = nil
			case "dmppatch":
				var buf string
				if len(lines) > 0 {
					buf = strings.Join(lines, "\n") + "\n"
				}
				err = apply(dir, []byte(buf), state, prevDiffInfo, curDiffInfo, dmpApply)
				if err != nil {
					return err
				}
				// reset
				prevDiffInfo = nil
				curDiffInfo = nil
			}
			state = fileDiff
		case terminal:
			return ErrNotTerminal
		default:
			return errors.New("patchfile: unknown state") // cannot happen
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	if state != terminal {
		return ErrPrematurePatchfileEnd
	}
	return nil
}
