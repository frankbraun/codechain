package hashchain

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/time"
)

const (
	chainStartType       = "cstart"
	sourceType           = "source"
	signatureType        = "signtr"
	addKeyType           = "addkey"
	removeKeyType        = "remkey"
	signatureControlType = "sigctl"
)

var emptyTree [32]byte

func init() {
	hash, err := hex.Decode(tree.EmptyHash, 32)
	if err != nil {
		panic(err)
	}
	copy(emptyTree[:], hash)
}

type link struct {
	previous   [32]byte // hash-of-previous
	datum      int64    // current-time
	linkType   string   // type
	typeFields []string // type-fields ...
}

func (l *link) String() string {
	return fmt.Sprintf("%x %s %s %s",
		l.previous,
		time.Format(l.datum),
		l.linkType,
		strings.Join(l.typeFields, " "))
}

func (l *link) Hash() [32]byte {
	return sha256.Sum256([]byte(l.String()))
}
