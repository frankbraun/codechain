/*
Package secpkg implements the secpkg package format.

A secure package (.secpkg file) contains a JSON object with the following keys:

  {
    "Name": "the project's package name",
    "Head": "head of project's Codechain",
    "DNS": "fully qualified domain name",
    "URL": "URL to download project files of the from (URL/head.tar.gz)"
  }

Example .secpkg file for Codechain itself:

  {
    "Name": "codechain",
    "Head": "73fe1313fd924854f149021e969546bce6052eca0c22b2b91245cb448410493c",
    "DNS": "codechain.secpkg.net",
    "URL": "http://frankbraun.org/codechain"
  }

Install specification

Installing software described by a .secpkg file works as follows:

   1. Parse .secpkg file and validate it. Save head (as HEAD_PKG).

   2. Make sure the project with NAME has not been installed before.
      That is, the directory ~/.config/secpkg/pkgs/NAME does not exist.

   3. Create directory ~/.config/secpkg/pkgs/NAME

   4. Save .secpkg file  to ~/.config/secpkg/pkgs/NAME/.secpkg

   5. Query TXT record from _codechain.DNS and validate the signed head
      contained in it (see ssot package). Save head from TXT record (HEAD_SSOT).

   6. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head.

   7. Download distribution file from URL/HEAD_SSOT.tar.gz and save it to
      ~/.config/secpkg/pkgs/NAME/dists

   8. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz
      to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
      -f ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz -head HEAD_SSOT`.

   9. Make sure HEAD_PKG is contained in
      ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain

  10. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`

  11. Call `make` in ~/.config/secpkg/pkgs/NAME/build.

  12. Call `make install` in ~/.config/secpkg/pkgs/NAME/build.

  If the installation process fails at any stage during the procedure described
  above, report the error and remove the directory ~/.config/secpkg/pkgs/NAME.

For the process above to work, the projects distributed as secure packages
must contain a Makefile (for GNU Make) with the "all" target building the
software and the "install" target installing it.

The software must be self-contained without any external dependencies, except
for the compiler. For Go software that means at least Go 1.11 must be
installed (with module support) and all dependencies must be vendored.

Update specification

TODO.
*/
package secpkg
