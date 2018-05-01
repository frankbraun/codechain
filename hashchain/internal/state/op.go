package state

import (
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/hashchain/linktype"
)

var nop = &noOP{}

type op interface {
	fmt.Stringer
	sign(w int)
	signatures() int
}

type signable struct {
	totalSignatures int
}

func (s *signable) sign(w int) {
	s.totalSignatures += w
}

func (s *signable) signatures() int {
	return s.totalSignatures
}

type noOP struct {
	signable
}

func (op *noOP) String() string {
	return ""
}

type sourceOP struct {
	signable
	treeHash string
	pubKey   string
	comment  string
}

func newSourceOP(treeHash, pubKey, comment string) *sourceOP {
	return &sourceOP{
		treeHash: treeHash,
		pubKey:   pubKey,
		comment:  comment,
	}
}

func (op *sourceOP) String() string {
	s := linktype.Source + " " + op.treeHash + " " + op.pubKey
	if op.comment != "" {
		s += " " + op.comment
	}
	return s
}

type addKeyOP struct {
	signable
	pubKey string
	weight int
}

func newAddKeyOP(pubKey string, weight int) *addKeyOP {
	return &addKeyOP{
		pubKey: pubKey,
		weight: weight,
	}
}

func (op *addKeyOP) String() string {
	return linktype.AddKey + " " + strconv.Itoa(op.weight) + " " + op.pubKey
}

type remKeyOP struct {
	signable
	pubKey string
}

func newRemKeyOP(pubKey string) *remKeyOP {
	return &remKeyOP{
		pubKey: pubKey,
	}
}

func (op *remKeyOP) String() string {
	return linktype.RemoveKey + " " + op.pubKey
}

type sigCtlOp struct {
	signable
	m int
}

func newSigCtlOp(m int) *sigCtlOp {
	return &sigCtlOp{
		m: m,
	}
}

func (op *sigCtlOp) String() string {
	return linktype.SignatureControl + " " + strconv.Itoa(op.m)
}
