/*
Package hashchain implements a hash chain of signatures over a chain of code
changes.

A hash chain is stored in a simple newline separated text file where each hash
chain entry corresponds to a line and has the following form:

  hash-of-previous current-time type type-fields ...

Where hash-of-previous is the SHA256 hash of the previous line (without the
trailing newline) in hex encoding. The fields are separated by single white
spaces. The current-time is encoded as an ISO 8601 string in UTC (corresponds
to time.RFC3339).

All hashes in a hash chain are SHA256 hashes encoded in hex notation. All
public keys are Ed25519 keys and they and their signatures are encoded in
base64 (URL encoding without padding). Comments are arbitrary UTF-8 sequences,
but cannot contain newlines.

Their are six different types of hash chain entries:

  cstart
  source
  signtr
  addkey
  remkey
  sigctl

A hash chain must start with a cstart entry and that is the only line where
this type must appear.


Type cstart

A cstart entry starts a new hash chain.

  hash-of-previous current-time cstart pubkey nonce signature [comment]

The hash-of-previous for the cstart time is the hash of an empty source tree
(see tree.EmptyHash). The signature by pubkey is over the pubkey, the nonce,
and the optional comment. The comment should identify the owner of the pubkey,
not the project. The nonce must be a 24 byte random number in base64 (URL
encoding without padding). This makes pubkey the only valid signer for the
hash chain and implicitly sets the signature threshold m to 1.


Type source

A source entry marks a new source tree state for publication from the
developer owning the signing pubkey. The optional comment can be used to
describe the change to the reviewers.

  hash-of-previous current-time source source-hash pubkey signature [comment]

The signature by pubkey is over the source-hash and the optional comment.
See the tree package for a detailed description of source tree hashes.


Type signtr

A signtr entry signs a previous hash chain entry and thereby approves all code
changes and changes to the set of signature keys and m up to that point.

  hash-of-previous current-time signtr hash-of-chain-entry pubkey signature

It does not (necessarily) sign the previous line and can therefore be done in
a detached fashion by a reviewer and added later by the developer responsible
for maintaining the hash chain. This avoids merge conflicts.


Type addkey

An addkey entry adds a signature pubkey to be added to the list of approved
signature keys.

  hash-of-previous current-time addkey pubkey-add w pubkey signature [comment]

The weight of the key towards the minimum number of necessary signatures m is
denoted by w. The pubkey can be accompanied by an optional comment, but the
signature must be over both. The comment is added last so it can contain white
spaces without complicating the parsing, it should identify the owner of the
pubkey.


Type remkey

A remkey entry marks a signature pubkey for removal.

  hash-of-previous current-time remkey pubkey


Type sigctl

A sigctl entry denotes an update of m, the minimum number of necessary
signatures to approve state changes (the threshold).

  hash-of-previous current-time sigctl m


Example

TODO
*/
package hashchain