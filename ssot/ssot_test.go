package ssot

import (
	"crypto/rand"
	"testing"

	"github.com/frankbraun/codechain/util/hex"
	"golang.org/x/crypto/ed25519"
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
	sh := SignHead(head, 0, sk)
	txt := sh.Marshal()
	_, err = Unmarshal(txt)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}
}
