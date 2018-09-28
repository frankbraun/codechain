Codechain — code trust through hash chains — β release
------------------------------------------------------

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/frankbraun/codechain) [![Build Status](https://img.shields.io/travis/frankbraun/codechain.svg?style=flat-square)](https://travis-ci.org/frankbraun/codechain) [![Go Report Card](https://goreportcard.com/badge/github.com/frankbraun/codechain?style=flat-square)](https://goreportcard.com/report/github.com/frankbraun/codechain)

This is a **beta** release of Codechain — everything is fairly stable,
but there is still a small risk that **the hash chain might be reset**!

### In code we trust: Secure multiparty code reviews with signatures and hash chains.

The most common signing mechanism for open-source software is using GPG
signatures. For example, GPG is used to sign Git commits and Debian
packages. There is no built-in mechanism for key rotation and key
compromise. And if forced to, a single developer can subvert all
machines which trust the corresponding GPG key.

That's where the Codechain tool comes in. It establishes code trust via
multi-party reviews recorded in unmodifiable hash chains.

Codechain allows to only publish code that has been reviewed by a
preconfigured set of reviewers. The signing keys can be rotated and the
reviewer set flexibly changed.

Every published code state is uniquely identified by a deterministic
source tree hash stored in the hash chain, signed by a single
responsible developer.

Codechain uses files to store the hash chain, not a distributed
"blockchain".

### Installation

```
go get -u -v github.com/frankbraun/codechain
```

### Features

- [x] Minimal code base, Go only, cross-platform.
- [ ] [Single source of truth (SSOT) with DNS](doc/ssot-with-dns.md)

Codechain depends on the `git` binary (for `git diff`), but that's optional.

### Out of scope

- Source code management. Git and other VCS systems are good for that,
  Codechain can be used alongside them and solves a different problem.
- Code distribution (minimal support is provided via `codechain
  createdist` and `codechain apply -f`).
- [Reproducible builds](https://reproducible-builds.org/).

### Documentation

- [Walkthrough](doc/walkthrough.md)
- [Presentation about Codechain](http://frankbraun.org/in-code-we-trust.pdf)
- [Hash chain file format](https://godoc.org/github.com/frankbraun/codechain/hashchain)
- [Patchfile format](https://godoc.org/github.com/frankbraun/codechain/patchfile)
- [Directory tree hashes and lists](https://godoc.org/github.com/frankbraun/codechain/tree)

### Acknowledgments

Codechain has been heavily influenced by discussions with
[Jonathan Logan](https://github.com/JonathanLogan) of
[Cryptohippie](https://secure.cryptohippie.com/), Inc.
