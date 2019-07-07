Bootstrapping Codechain
-----------------------

THIS DOCUMENT IS NOT FINISHED YET!

-   [Install Go](https://golang.org/doc/install).
-   Download `cctreehash` from
    https://raw.githubusercontent.com/frankbraun/codechain/master/cmd/util/cctreehash/cctreehash.go
    and verify its SHA-256:

    \$ sha256sum cctreehash.go
    087076108dc3366be27d42348fb846a591bf994c7d5e90ceba2b0ad2ab9f8dfe
    cctreehash.go

-   If you can, verify the code of `cctreehash.go`.
-   [Download Codechain
    bootstrap](https://frankbraun.org/codechain-bootstrap.tar.gz) and
    execute:

    \$ tar -xvf codechain-bootstrap.tar.gz \$ cd codechain-bootstrap \$
    go run ../cctreehash.go
    d1e69edf8f1c09e82fd16b008d70ec0783982418a4f113f19f661a7b47919641

    The tree hash computed by `cctreehash.go` must match the published
    one. This makes sure you have the correct bootstrap Codechain
    source.

-   Use the bootstrapped Codechain to install the most current Codechain
    version:

    go run cmd/secpkg/secpkg.go install .secpkg
