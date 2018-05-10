package patchfile

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func dmpDiff(w io.Writer, fileA, fileB string) error {
	var a []byte
	if fileA != "" {
		var err error
		a, err = ioutil.ReadFile(fileA)
		if err != nil {
			return err
		}
	}
	b, err := ioutil.ReadFile(fileB)
	if err != nil {
		return err
	}
	dmp := diffmatchpatch.New()
	textA, textB, lineArray := dmp.DiffLinesToRunes(string(a), string(b))
	diffs := dmp.DiffMainRunes(textA, textB, true)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)
	patches := dmp.PatchMake(string(a), diffs)
	patch := dmp.PatchToText(patches)
	/*
		if !strings.HasSuffix(patch, "\n") {
			patch = patch + "\n"
		}
	*/
	fmt.Fprintf(w, "dmppatch %d\n", strings.Count(patch, "\n"))
	if _, err := io.WriteString(w, patch); err != nil {
		return err
	}

	dmp = diffmatchpatch.New()
	patches, err = dmp.PatchFromText(patch)
	if err != nil {
		return err
	}
	_, applies := dmp.PatchApply(patches, string(a))
	for _, applied := range applies {
		if !applied {
			return errors.New("patchfile: XXX could not apply all patches")
		}
	}

	return nil
}

func dmpApply(w io.Writer, text string, patch []byte) error {
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(string(patch))
	if err != nil {
		return err
	}
	newText, applies := dmp.PatchApply(patches, text)
	for _, applied := range applies {
		if !applied {
			return errors.New("patchfile: could not apply all patches")
		}
	}
	if _, err := io.WriteString(w, newText); err != nil {
		return err
	}
	return nil
}
