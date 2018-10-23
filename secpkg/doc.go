/*
Package secpkg implements the secpkg package format.

A secure package (.secpkg file) contains a JSON object with the following keys:

  {
    "Name": "the project's package name",
    "Head": "head of project's Codechain",
    "DNS": "fully qualified domain name for Codechain's TXT records",
  }

Example .secpkg file for Codechain itself:

  {
    "Name": "codechain",
    "Head": "53f2c26d92e173306e83d54e3103ef2e0bd87a561315bc4b49e1ee6c78dfb583",
    "DNS": "codechain.secpkg.net",
  }

Install specification

Installing software described by a .secpkg file works as follows:

   1. Parse .secpkg file and validate it. Save head as HEAD_PKG.

   2. Make sure the project with NAME has not been installed before.
      That is, the directory ~/.config/secpkg/pkgs/NAME does not exist.

   3. Create directory ~/.config/secpkg/pkgs/NAME

   4. Save .secpkg file to ~/.config/secpkg/pkgs/NAME/.secpkg

   5. Query TXT record from _codechain-head.DNS and validate the signed head
      contained in it (see ssot package). Save head from TXT record (HEAD_SSOT).

   6. Query TXT record from _codechain-url.DNS and save it as URL.

   7. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head

   8. Download distribution file from URL/HEAD_SSOT.tar.gz and save it to
      ~/.config/secpkg/pkgs/NAME/dists

   9. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz
      to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
      -f ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.tar.gz -head HEAD_SSOT`

  10. Make sure HEAD_PKG is contained in
      ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain

  11. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`

  12. Call `make prefix=~/.config/secpkg/local` in
      ~/.config/secpkg/pkgs/NAME/build

  14. Call `make prefix= ~/.config/secpkg/local install` in
      ~/.config/secpkg/pkgs/NAME/build

  15. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`

  If the installation process fails at any stage during the procedure described
  above, report the error and remove the directory ~/.config/secpkg/pkgs/NAME.

For the process above to work, the projects distributed as secure packages
must contain a Makefile (for GNU Make) with the "all" target building the
software and the "install" target installing it.

The software must be self-contained without any external dependencies, except
for the compiler. For Go software that means at least Go 1.11 must be
installed (with module support) and all dependencies must be vendored.

Update specification

Updating a software package with NAME works as follows:

   1. Make sure the project with NAME has been installed before.
      That is, the directory ~/.config/secpkg/pkgs/NAME exists.

   2. Load .secpkg file from ~/.config/secpkg/pkgs/NAME/.secpkg

   3. Load signed head from ~/.config/secpkg/pkgs/NAME/signed_head (as DISK)

   4. Query TXT record from _codechain-head.DNS, if it is the same as DISK, goto 16.

   5. Query TXT record from _codechain-url.DNS and save it as URL.

   6. Validate signed head from TXT (also see ssot package) and store HEAD:

      - pubKey from TXT must be the same as pubKey or pubKeyRotate from DISK.
      - The counter from TXT must be larger than the counter from DISK.
      - The signed head must be valid (as defined by validFrom and validTo).

      If the validation fails, abort update procedure and report error.

   7. If signed head from TXT record is the same as the one from DISK:

      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
               ~/.config/secpkg/pkgs/NAME/previous_signed_head`
      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).
      - Goto 16.

   8. Download distribution file from URL/HEAD.tar.gz and save it to
      ~/.config/secpkg/pkgs/NAME/dists

   9. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz
      to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
      -f ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz -head HEAD`.

  10. `rm -rf ~/.config/secpkg/pkgs/NAME/build`

  11. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`

  12. Call `make prefix=~/.config/secpkg/local` in
      ~/.config/secpkg/pkgs/NAME/build

  13. Call `make prefix= ~/.config/secpkg/local install` in
      ~/.config/secpkg/pkgs/NAME/build

  14. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`

  15. Update signed head:

      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head
               ~/.config/secpkg/pkgs/NAME/previous_signed_head`
      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head (atomic).

  16. The software has been successfully updated.
*/
package secpkg
