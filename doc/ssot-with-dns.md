Single source of truth (SSOT) with DNS notes
--------------------------------------------

### TXT records

We have to store the following information:

-   pubkey (32 byte)
-   pubkeyRotate (32 byte)
-   validity: from, to (16 byte)
-   counter (8 byte)
-   head (32 byte)
-   signature (64 byte)

The protocol shall define a global maximum validity.

Implement:

-   export in Bind format
-   resolver for records
-   optional: automatic publishing with
    [lego](https://github.com/xenolf/lego)?

### Secure packages (`.secpkg`)

See [secpkg package
format](https://godoc.org/github.com/frankbraun/codechain/secpkg).

### DNSSEC

TODO:

-   TLD should allow DNSSEC (all of them?)
-   Registar should support DNSSEC
-   DNSSEC should be activated

### Possible attacks

-   publisher attack: not possible
-   DNS poisoning:
    -   user saw key before: failed
    -   user didn't see key before: success (can be mitigated with
        DNSSEC)

### Results

This gives us

-   globally identical,
-   verifiable,
-   reproducible, and
-   attributable

Go binaries!
