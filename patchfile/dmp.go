package patchfile

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/frankbraun/go-diff/diffmatchpatch"
)

// Can panic!
func diff(a, b string) string {
	dmp := diffmatchpatch.New()
	textA, textB, lineArray := dmp.DiffLinesToRunes(a, b)
	diffs := dmp.DiffMainRunes(textA, textB, true)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)
	patches := dmp.PatchMake(a, diffs)
	return dmp.PatchToText(patches)
}

// dmpDiff employs Myers' diff algorithm (as implemented in Diff Match Patch)
// to calculate a diff between fileA and fileB, and writes it to w as a
// "dmppatch" section.
//
// If dmpDiff was able to calculate a diff that will apply cleanly, it returns
// true. Otherwise, it returns false.
func dmpDiff(w io.Writer, fileA, fileB string) (bool, error) {
	var a []byte
	if fileA != "" {
		var err error
		a, err = ioutil.ReadFile(fileA)
		if err != nil {
			return false, err
		}
	}
	b, err := ioutil.ReadFile(fileB)
	if err != nil {
		return false, err
	}
	aStr := string(a)
	bStr := string(b)
	var (
		patch    string
		panicked bool
	)
	// call diff() which can panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		patch = diff(aStr, bStr)
	}()
	// diff() panicked, we cannot compute a clean diff
	if panicked {
		return false, nil
	}
	// make sure patch applies cleanly
	text, err := patchApply(aStr, patch)
	if err != nil || text != bStr {
		return false, nil
	}
	// all good, write patch
	fmt.Fprintf(w, "dmppatch %d\n", strings.Count(patch, "\n"))
	if _, err := io.WriteString(w, patch); err != nil {
		return false, err
	}
	return true, nil
}

func patchApply(text, patch string) (string, error) {
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(patch)
	if err != nil {
		return "", err
	}
	newText, applies := dmp.PatchApply(patches, text)
	for _, applied := range applies {
		if !applied {
			return "", errors.New("patchfile: could not apply all patches")
		}
	}
	return newText, nil
}

// dmpApply decodes the DMP patch in patch, applies it to text, and writes it
// to w. patch must not include the "dmppatch" section header.
func dmpApply(w io.Writer, text string, patch []byte) error {
	newText, err := patchApply(text, string(patch))
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, newText); err != nil {
		return err
	}
	return nil
}
