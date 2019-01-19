package hashchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/time"
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

func linkEqual(a, b *link) bool {
	if !bytes.Equal(a.previous[:], b.previous[:]) {
		return false
	}
	if a.datum != b.datum {
		return false
	}
	if a.linkType != b.linkType {
		return false
	}
	if len(a.typeFields) != len(b.typeFields) {
		return false
	}
	for i, field := range a.typeFields {
		if field != b.typeFields[i] {
			return false
		}
	}
	return true
}

func (l *link) String() string {
	return fmt.Sprintf("%x %s %s %s",
		l.previous,
		time.Format(l.datum),
		l.linkType,
		strings.Join(l.typeFields, " "))
}

func (l *link) StringColor() string {
	// hash-of-previous / hash-of-chain-entry: green
	// current-time: white
	// type: black
	// pubkey: red
	// nonce: magenta
	// signature: blue
	// source-hash: cyan
	// w/m: HiRed
	// comment: yellow
	s := color.GreenString("%x", l.previous) + " " +
		color.WhiteString(time.Format(l.datum)) + " " +
		l.linkType + " "
	switch l.linkType {
	case "cstart":
		s += color.RedString(l.typeFields[0]) + " " +
			color.MagentaString(l.typeFields[1]) + " " +
			color.BlueString(l.typeFields[2])
		if len(l.typeFields) == 4 {
			s += " " + color.YellowString(l.typeFields[3])
		}
	case "source":
		s += color.CyanString(l.typeFields[0]) + " " +
			color.RedString(l.typeFields[1]) + " " +
			color.BlueString(l.typeFields[2])
		if len(l.typeFields) == 4 {
			s += " " + color.YellowString(l.typeFields[3])
		}
	case "signtr":
		s += color.GreenString(l.typeFields[0]) + " " +
			color.RedString(l.typeFields[1]) + " " +
			color.BlueString(l.typeFields[2])
	case "addkey":
		s += color.HiRedString(l.typeFields[0]) + " " +
			color.RedString(l.typeFields[1]) + " " +
			color.BlueString(l.typeFields[2])
		if len(l.typeFields) == 4 {
			s += " " + color.YellowString(l.typeFields[3])
		}
	case "remkey":
		s += color.RedString(l.typeFields[0])
	case "sigctl":
		s += color.HiRedString(l.typeFields[0])
	default:
		panic("hashchain: unknown link type")
	}
	return s
}

func (l *link) Hash() [32]byte {
	return sha256.Sum256([]byte(l.String()))
}
