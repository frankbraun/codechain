Bootstrapping Codechain
-----------------------

THIS DOCUMENT IS NOT FINISHED YET!

### Install Go

See https://golang.org/doc/install.

### Download `cctreehash.go`

From https://frankbraun.org/cctreehash.go and verify its SHA-256:

    $ sha256sum cctreehash.go
    fef870afcb3b0f903e71b153c29df7c20f63a10d838dff08e55d8b2c6bf9a2e9 cctreehash.go

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

### Use the bootstrapped Codechain to install the most current Codechain version

    $ go run cmd/secpkg/secpkg.go install .secpkg

Afterwards `codechain` and `secpkg` are installed in
`~/.config/secpkg/local/bin`. You should add that directory to your
`$PATH` variable and delete the `codechain-bootstrap` directory.
