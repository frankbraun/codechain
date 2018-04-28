package hashchain

import (
	"encoding/hex"
	"fmt"
	"strings"

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

var emptyTree []byte

func init() {
	var err error
	emptyTree, err = hex.DecodeString(tree.EmptyHash)
	if err != nil {
		panic(err)
	}
}

type link struct {
	previous   []byte   // hash-of-previous
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
