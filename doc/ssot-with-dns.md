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
-   optional: automatic publishing with [lego](https://github.com/xenolf/lego)?

### Secure packages (`.secpkg`)

File format:

    {
      "Name": "the project's package name",
      "Head": "head of project's Codechain",
      "DNS": "fully qualified domain name",
      "URL": "URL to download project files of the from (URL/head.tar.gz)"
    }

Example:

    {
      "Name": "codechain",
      "Head": "73fe1313fd924854f149021e969546bce6052eca0c22b2b91245cb448410493c",
      "DNS": "codechain.secpkg.net",
      "URL": "http://frankbraun.org/codechain"
    }

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
