package patchfile

import (
	//"io/ioutil"
	//"path/filepath"
	"testing"
	//"github.com/frankbraun/codechain/internal/ascii85"
)

// Tests that show where github.com/sergi/go-diff/diffmatchpatch fails.
func TestDMP(t *testing.T) {
	// TODO: enable test
	/*
		filename := filepath.Join("testdata", "dmp", "tables.go.bin")
		enc, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("ioutil.ReadFile() failed: %v", err)
		}
		// tables.go.bin -> tables.go
		b, err := ascii85.Decode(enc)
		if err != nil {
			t.Fatalf("ascii85.Decode() failed: %v", err)
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
	*/
}
