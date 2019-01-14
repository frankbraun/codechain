package patchfile

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// Tests that show where github.com/sergi/go-diff/diffmatchpatch fails.
func TestDMP(t *testing.T) {
	filename := filepath.Join("testdata", "dmp", "tables.go")
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	// patchApply fails
	text := string(b)
	patch := diff("", text)
	newText, err := patchApply("", patch)
	if err != nil {
		t.Fatalf("patchApply() failed: %v", err)
	}
	if text == newText {
		t.Error("newText equals text (DMP should fail here")
	}
	// diff panics
	text = string(b[420000:])
	defer func() {
		if x := recover(); x == nil {
			t.Error("diff() should panic")
		}
	}()
	patch = diff("", text)
}
