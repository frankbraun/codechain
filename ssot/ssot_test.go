package ssot

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/frankbraun/codechain/util/hex"
)

const headStr = "73fe1313fd924854f149021e969546bce6052eca0c22b2b91245cb448410493c"

func TestSignedHead(t *testing.T) {
	buf, err := hex.Decode(headStr, 32)
	if err != nil {
		t.Fatalf("hex.Decode() failed: %v", err)
	}
	var head [32]byte
	copy(head[:], buf)
	_, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() failed: %v", err)
	}
	var sk [64]byte
	copy(sk[:], sec)

	// error cases
	_, err = SignHead(head, 0, sk, nil, MinimumValidity-time.Second)
	if err != ErrValidityTooShort {
		t.Error("SignHead() should fail with ErrValidityTooShort")
	}
	_, err = SignHead(head, 0, sk, nil, MaximumValidity+time.Second)
	if err != ErrValidityTooLong {
		t.Error("SignHead() should fail with ErrValidityTooLong")
	}

	// happy cases
	_, err = SignHead(head, 0, sk, nil, MinimumValidity)
	if err != nil {
		t.Fatalf("SignHead() failed: %v", err)
	}
	sh, err := SignHead(head, 0, sk, nil, MaximumValidity)
	if err != nil {
		t.Fatalf("SignHead() failed: %v", err)
	}
	txt := sh.Marshal()
	_, err = Unmarshal(txt)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// V2
	sh2, err := SignHeadV2(head, 2, 0, sk, nil, MinimumValidity)
	if err != nil {
		t.Fatalf("SignHead() failed: %v", err)
	}
	if len(sh2.Marshal()) > 255 {
		t.Error("V2 head is too long")
	}
	/*
		txt = sh2.Marshal()
		_, err = Unmarshal(txt)
		if err != nil {
			t.Fatalf("Unmarshal() failed: %v", err)
		}
	*/
}
