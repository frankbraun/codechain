/*
Package hashchain implements a hash chain of signatures over a chain of code
changes.

A hash chain is stored in a simple newline separated text file where each hash
chain entry corresponds to a single line and has the following form:

  hash-of-previous current-time type type-fields ...

Where hash-of-previous is the SHA256 hash of the previous line (without the
trailing newline) in hex encoding. The fields are separated by single white
spaces. The current-time is encoded as an ISO 8601 string in UTC (corresponds
to time.RFC3339).

All hashes in a hash chain are SHA256 hashes encoded in hex notation.
Hex encodings have to be lowercase. All public keys are Ed25519 keys and they
and their signatures are encoded in base64 (URL encoding without padding).
Comments are arbitrary UTF-8 sequences, but cannot contain newlines.

There are six different types of hash chain entries:

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

  hash-of-previous current-time source tree-hash pubkey signature [comment]

The signature by pubkey is over the source tree hash and the optional comment.
See the tree package for a detailed description of source tree hashes.


Type signtr

A signtr entry signs a previous hash chain entry and thereby approves all code
changes and changes to the set of signature keys and m up to that point.

  hash-of-previous current-time signtr hash-of-chain-entry pubkey signature

It does not necessarily sign the previous line and can therefore be done in a
detached fashion by a reviewer and added later by the developer responsible
for maintaining the hash chain. This avoids merge conflicts.


Type addkey

An addkey entry marks a signature pubkey for addition to the list of approved
signature keys.

  hash-of-previous current-time addkey w pubkey signature [comment]

The weight of the key towards the minimum number of necessary signatures m is
denoted by w. The pubkey can be accompanied by an optional comment, but the
signature must be over both. The comment is added last so it can contain white
spaces without complicating the parsing, it should identify the owner of the
pubkey.


Type remkey

A remkey entry marks a signature pubkey for removal from the list of approved
signature keys.

  hash-of-previous current-time remkey pubkey


Type sigctl

A sigctl entry denotes an update of m, the minimum number of necessary
signatures to approve state changes (the threshold).

  hash-of-previous current-time sigctl m


Example

An example of a hash chain.

  e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 2018-05-19T00:07:02Z cstart KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 sVnVenzHyCOV6nLUkCKg6ARllkYsTV-n 0UmUcDFZ2j3WWnqzEdxX-wzofWlhF3O0Rm1tT6qMUwLu8a1R5MwbK5zDongYZKccpA37Vp6Sp3m0xSreGskzCg Alice <alice@example.com>
  40c7e5ca4be98e9cae6931afa4ac09e11ecb1ce20fa18d0faaabfac7e8fad071 2018-05-19T00:09:44Z addkey 1 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc Xsr_L-1_5_B56vocve8s3Pb3vJoc-jpa2-tzIQhEjuoytYfcAiONu3er6RnVNMcsPuZFeqWCQKBwka-F-c13Ag Bob <bob@example.com>
  34cd10effd93e67ba96fefb29ea751d013459a6de11cc117cf1deacd77d6b7be 2018-05-19T00:10:25Z sigctl 2
  92d2fc6687b0d36d045adaf34a1615e513ef0e2dc60384cfe19863e9753567f8 2018-05-19T00:11:44Z source d844cbe6f6c2c29e97742b272096407e4d92e6ac7f167216b321c7aa55629716 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 r5aZCYGwWCFppaMDV7XSOHoyCl3qbUKGiSuYzjsTl4C0W9n0tCa0MXDy_fOwspV9f4_o0kMcb6XZS706ml3FAQ first release
  d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 2018-05-19T00:12:48Z signtr d258ce20943beeed2d483096702a1449447f112dec7d907d50c285c649c17a24 KDKOGoY8ErjOnbDQb4k8SZFMvWdAIb-x6FGKKCRby70 HKlLKnYSCVzc4b-erETK50EN5gKRKZQsT16grv7eFBklFqXBFoSXSmcY99HLWhAP9BJcA6c3Px1trNBns3KkDA
  2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 2018-05-19T00:34:51Z signtr 2e34e23ee293e8c0ed174639d325eb3e30f5337d5c5846380367724e93cb619e 91HOu2fvkjHd5S0LtAWTl6dYBk5cqB-NWiJqc0c_7Gc xffZultos-MCbI4cNzAzAoccuDSnpL2nq_BsQanIruYM3RXoD9kdC6WiPEUkxrphKdG742IgBWlB3LwY0i1ZCw
*/
package hashchain
