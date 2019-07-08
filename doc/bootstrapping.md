Bootstrapping Codechain
-----------------------

THIS DOCUMENT IS NOT FINISHED: The file codechain-bootstrap.tar.gz has
not been published and the tree hash has not been signed yet.

### Install Go

See https://golang.org/doc/install.

### Download `cctreehash.go`

From https://frankbraun.org/cctreehash.go and verify its SHA-256:

    $ sha256sum cctreehash.go
    6a911b51cc047eea7d7b10c75cf1f50a98dea2a39e12ef3414d2f00f31a263f9 cctreehash.go

### Review the code of `cctreehash.go`

If you can, this is optional.

### Download Codechain bootstrap

From https://frankbraun.org/codechain-bootstrap.tar.gz and execute:

    $ tar -xvf codechain-bootstrap.tar.gz
    $ cd codechain-bootstrap
    $ go run ../cctreehash.go
    d1e69edf8f1c09e82fd16b008d70ec0783982418a4f113f19f661a7b47919641

The tree hash computed by `cctreehash.go` must match the published one.
This makes sure you have the correct Codechain bootstrap source.

TODO: sources of the `codechain-bootrap` tree hash:

-   opentimestamps
-   PGP signed statements by developers
-   websites from projects that use Codechain

### Use the bootstrapped Codechain to install the most current Codechain version

    $ go run cmd/secpkg/secpkg.go install .secpkg

Afterwards `codechain` and `secpkg` are installed in
`~/.config/secpkg/local/bin`. You should add that directory to your
`$PATH` variable and delete the `codechain-bootstrap` directory.
