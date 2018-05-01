// Package linktype defines the different link types of a hash chain.
package linktype

const (
	// ChainStart link type.
	ChainStart = "cstart"
	// Source link type.
	Source = "source"
	// Signature link type.
	Signature = "signtr"
	// AddKey link type.
	AddKey = "addkey"
	// RemoveKey link type.
	RemoveKey = "remkey"
	// SignatureControl link type.
	SignatureControl = "sigctl"
)
