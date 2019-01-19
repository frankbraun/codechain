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

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
)

// state machine used to apply patch files.
type state int

// all possible state:
const (
	start          state = iota + 1 // start state
	treehash                        // tree hash
	fileDiff                        // first file diff
	secondFileDiff                  // second file diff
	addFile                         // add file state
	diffFile                        // diff file state
	terminal                        // end state
)

// codechain patchfile version
func procStart(line string) (state, int, error) {
	fields := strings.SplitN(line, " ", 4)
	if len(fields) != 4 {
		return 0, 0, ErrHeaderFieldsNum
	}
	if fields[0] != "codechain" || fields[1] != "patchfile" || fields[2] != "version" {
		return 0, 0, ErrHeaderFieldsText
	}
	version, err := strconv.Atoi(fields[3])
	if err != nil {
		return 0, 0, err
	}
	if version != 1 && version != 2 { // only support Version 1 and 2
		return 0, 0, ErrHeaderVersion
	}
	return treehash, version, nil
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
		// We read to two file info lines after another ('-' followed by '+').
		if name != prevDiffInfo.name {
			// The file names changed, we have to "move" the file.
			// First make sure target doesn't exist.
			fn := filepath.Join(dir, name)
			exists, err := file.Exists(fn)
			if err != nil {
				return 0, nil, err
			}
			if exists {
				return 0, nil, ErrMoveTargetFileExists
			}
			// Instead of "moving", we delete the previous file and then add
			// the current file again.
			oldpath := filepath.Join(dir, prevDiffInfo.name)
			if err := os.Remove(oldpath); err != nil {
				return 0, nil, err
			}
			return addFile, &diffInfo{mode, hash, name}, nil
		}
		// The two files have the same name, check if their hash differs.
		if bytes.Equal(hash[:], prevDiffInfo.hash) {
			// The hashes do not differ, check if we have to change the file
			// permissions.
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
			// Because the files have the same names and the same hash we do
			// not process a diff next, but expect another file line.
			return fileDiff, nil, nil
		}
		// The hash differs, we have to process a diff next.
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

// procResults process the final treehash line and returns the terminal state.
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
	// If the file already existed os.OpenFile didn't change the file mode,
	// so we adjust it here.
	if state == diffFile && prev.mode != cur.mode {
		if err := os.Chmod(fileB, perm); err != nil {
			return err
		}
	}
	return nil
}

type diffInfo struct {
	mode
	hash []byte
	name string
}

// scanNewlines is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may be
// empty. The end-of-line marker is one mandatory newline. In regular
// expression notation, it is `\n`. The last non-empty line of input will be
// returned even if it has no newline.
func scanNewlines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil

	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Apply applies the patch read from r to the directory tree dir.
// The paths given in excludePaths are excluded from all tree hash calculations.
func Apply(dir string, r io.Reader, excludePaths []string) error {
	log.Println("patchfile.Apply()")
	var (
		prevDiffInfo *diffInfo
		curDiffInfo  *diffInfo
		err          error
	)
	s := bufio.NewScanner(r)
	buf := make([]byte, bufio.MaxScanTokenSize)
	s.Buffer(buf, 64*1024*1024) // 64MB, entire files can be encoded as single lines
	s.Split(scanNewlines)
	state := start
	version := 0
	for s.Scan() {
		line := s.Text()
		log.Println("line:")
		log.Println(line)
		switch state {
		case start:
			log.Println("state: start")
			state, version, err = procStart(line)
			if err != nil {
				return err
			}
		case treehash:
			log.Println("state: treehash")
			state, err = procTreeHash(line, dir, excludePaths)
			if err != nil {
				return err
			}
		case fileDiff:
			log.Println("state: fileDiff")
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
			log.Println("state: secondFileDiff")
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
			log.Println("state: addFile (fallthrough)")
			fallthrough
		case diffFile:
			log.Println("state: diffFile")
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
			case "utf8file":
				if version < 2 {
					return ErrDiffModeUnknown
				}
				if numLines < 1 {
					return ErrDiffLinesNonPositive
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
			case "utf8file":
				buf := strings.Join(lines, "\n")
				err = apply(dir, []byte(buf), state, prevDiffInfo, curDiffInfo, utf8fileApply)
				if err != nil {
					return err
				}
				// reset
				prevDiffInfo = nil
				curDiffInfo = nil
			}
			state = fileDiff
		case terminal:
			log.Println("state: terminal")
			return ErrNotTerminal
		default:
			log.Println("state: unknown")
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
