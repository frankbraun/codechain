package ascii85

import (
	"bytes"
	"testing"

	"github.com/frankbraun/codechain/internal/hex"
)

const (
	b64 = "21543db19f5b682f3d07bdacef6b2c31804021b7b98fbe196d9d4d828df16270" +
		"70a12f80717235e6aa48111187e3ddd935e09b61bd0289e0fd08ad748f9a39af"
)

func TestEncodeDecode64(t *testing.T) {
	buf, err := hex.Decode(b64, 64)
	if err != nil {
		t.Fatalf("hex.Decode() failed: %v", err)
	}

	enc, lines, err := Encode(buf)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	if lines != 1 {
		t.Error("lines should equal 1")
	}

	dec, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !bytes.Equal(dec, buf) {
		t.Error("Encode() + Decode() failed")
	}
}

func TestEncodeDecode65(t *testing.T) {
	buf, err := hex.Decode(b64+"ff", 65)
	if err != nil {
		t.Fatalf("hex.Decode() failed: %v", err)
	}

	enc, lines, err := Encode(buf)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	if lines != 2 {
		t.Error("lines should equal 2")
	}

	dec, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !bytes.Equal(dec, buf) {
		t.Error("Encode() + Decode() failed")
	}
}
