package tree

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
)

const testdataList = `d 755 bar
f 644 7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730 bar/baz.txt
f 644 b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c foo.txt
`

const testdataHash = "9644253f9bd469f4771ff085afac826a17b5690b09c376bee2daf26ecd199d50"

func TestEmpty(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "tree_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	l, err := List(tmpdir, nil)
	if err != nil {
		t.Errorf("List() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte{}) {
		t.Errorf("List() should return an empty list")
	}
	h, err := Hash(tmpdir, nil)
	if err != nil {
		t.Errorf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h) != EmptyHash {
		t.Errorf("Hash() should return the EmptyHash")
	}
}

func TestTestdata(t *testing.T) {
	l, err := List("testdata", nil)
	if err != nil {
		t.Errorf("List() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte(testdataList)) {
		t.Errorf("List() should return testdataList")
	}
	h, err := Hash("testdata", nil)
	if err != nil {
		t.Errorf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h) != testdataHash {
		t.Errorf("Hash() should return the testdataHash")
	}
}

func TestTestdataChdir(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		t.Errorf("os.Chdir() should not fail: %v", err)
	}
	l, err := List(".", nil)
	if err != nil {
		t.Errorf("List() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte(testdataList)) {
		t.Errorf("List() should return testdataList")
	}
	h, err := Hash(".", nil)
	if err != nil {
		t.Errorf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h) != testdataHash {
		t.Errorf("Hash() should return the testdataHash")
	}
}
