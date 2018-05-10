package tree

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
)

func TestEmpty(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "tree_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	l, err := ListBytes(tmpdir, nil)
	if err != nil {
		t.Fatalf("ListBytes() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte{}) {
		t.Errorf("ListBytes() should return an empty list")
	}
	h, err := Hash(tmpdir, nil)
	if err != nil {
		t.Fatalf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h[:]) != EmptyHash {
		t.Errorf("Hash() should return the EmptyHash")
	}
}

const testdataList = `f 7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730 bar/baz.txt
f b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c foo.txt
`

const testdataHash = "6a311afb95a38e3de2cec9a8566d637198097abe734fef698d8032b80272dc1b"

func TestTestdata(t *testing.T) {
	l, err := ListBytes("testdata", nil)
	if err != nil {
		t.Errorf("ListBytes() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte(testdataList)) {
		t.Errorf("ListBytes() should return testdataList")
	}
	h, err := Hash("testdata", nil)
	if err != nil {
		t.Errorf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h[:]) != testdataHash {
		t.Errorf("Hash() should return the testdataHash")
	}
}

const testdataListExclude = `f 7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730 bar/baz.txt
`
const testdataHashExclude = "45e169478843023bb6d8c7fdcca1f1b199404975e2ebb09e76df6819a5dcad10"

func TestTestdataExclude(t *testing.T) {
	excludePaths := []string{"foo.txt"}
	l, err := ListBytes("testdata", excludePaths)
	if err != nil {
		t.Fatalf("ListBytes() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte(testdataListExclude)) {
		t.Errorf("ListBytes() should return testdataListExclude")
	}
	h, err := Hash("testdata", excludePaths)
	if err != nil {
		t.Fatalf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h[:]) != testdataHashExclude {
		t.Errorf("Hash() should return the testdataHashExclude")
	}
}

func TestTestdataChdir(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		t.Fatalf("os.Chdir() should not fail: %v", err)
	}
	l, err := ListBytes(".", nil)
	if err != nil {
		t.Fatalf("ListBytes() should not fail: %v", err)
	}
	if !bytes.Equal(l, []byte(testdataList)) {
		t.Errorf("ListBytes() should return testdataList")
	}
	h, err := Hash(".", nil)
	if err != nil {
		t.Fatalf("Hash() should not fail: %v", err)
	}
	if hex.EncodeToString(h[:]) != testdataHash {
		t.Errorf("Hash() should return the testdataHash")
	}
}
