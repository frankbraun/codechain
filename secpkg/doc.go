/*
Package secpkg implements the secpkg package format.

A secure package (.secpkg file) contains a JSON object with the following
mandatory keys:

  {
    "Name": "the project's package name",
    "Head": "head of project's Codechain",
    "DNS": "fully qualified domain name for Codechain's TXT records"
  }

It can contain multiple optional secondary domain names:

  {
    "Name": "the project's package name",
    "Head": "head of project's Codechain",
    "DNS": "fully qualified domain name for Codechain's TXT records",
    "DNS2": [
      "secondary fully qualified domain name for Codechain's TXT records"
    ]
  }

Example .secpkg file for Codechain itself:

  {
    "Name": "codechain",
    "Head": "53f2c26d92e173306e83d54e3103ef2e0bd87a561315bc4b49e1ee6c78dfb583",
    "DNS": "codechain.secpkg.net",
  }

If in the root of a package source tree the directory .secdep exists and
contains any .secpkg files, then these secure dependencies are installed and
kept up-to-date by the install and update procedures specified below.

In the following, the DNS entry and the DNS2 entries combined in random order
are called DNS_RECORDS. Since a .secpkg file must contain a DNS entry
DNS_RECORDS contains at least one entry.

Install specification

Installing software described by a .secpkg file works as follows:

   1. Parse .secpkg file and validate it. Save head as HEAD_PKG.

   2. Make sure the project with NAME has not been installed before.
      That is, the directory ~/.config/secpkg/pkgs/NAME does not exist.

   3. Create directory ~/.config/secpkg/pkgs/NAME

   4. Save .secpkg file to ~/.config/secpkg/pkgs/NAME/.secpkg

   5. Get next DNS entry from DNS_RECORDS. If no such entry exists: Goto 8.

   6. Query TXT record from _codechain-head.DNS and validate the signed head
      contained in it (see ssot package). Save head from TXT record
      (HEAD_SSOT.DNS).

   7. Store the signed head to ~/.config/secpkg/pkgs/NAME/signed_head.DNS
      Goto 5.

   8. Sort DNS_RECORDS in descending order according to the last signed line
      number (signed head version 2 or higher).

   9. Get next DNS entry from DNS_RECORDS. If no such entry exists, exit with
      error.

  10. Query all TXT records from _codechain-url.DNS and save it as URLs.

  11. Select next URL from URLs. If no such URL exists: Goto 9.

  12. Download distribution file from URL/HEAD_SSOT.DNS.tar.gz and save it to
      ~/.config/secpkg/pkgs/NAME/dists
      If it fails: Goto 11.

  13. Apply ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.DNS.tar.gz
      to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
      -f ~/.config/secpkg/pkgs/NAME/dists/HEAD_SSOT.DNS.tar.gz
      -head HEAD_SSOT.DNS`
      If it fails: Goto 11.

  14. Make sure HEAD_PKG is contained in
      ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain
      If it fails: Goto 11.

  15. If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
      contains any .secpkg files, ensure these secure dependencies are
      installed and up-to-date.

  16. `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`

  17. Call `make prefix=~/.config/secpkg/local` in
      ~/.config/secpkg/pkgs/NAME/build

  18. Call `make prefix= ~/.config/secpkg/local install` in
      ~/.config/secpkg/pkgs/NAME/build

  19. `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`

  20. If the file ~/.config/secpkg/pkgs/NAME/installed/.secpkg exists,
      `cp -f ~/.config/secpkg/pkgs/NAME/installed/.secpkg
             ~/.config/secpkg/pkgs/NAME/.secpkg`

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

   3. Get next DNS entry from DNS_RECORDS. Set SKIP_BUILD.DNS to false.
      If no such entry exists: Goto 7.

   4. Load signed head from ~/.config/secpkg/pkgs/NAME/signed_head.DNS
      (as DISK.DNS)

   5. Query TXT record from _codechain-head.DNS, if it is the same as DISK.DNS,
      set SKIP_BUILD.DNS to true.

   6. Update signed head:

      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head.DNS
               ~/.config/secpkg/pkgs/NAME/previous_signed_head.DNS`
      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head.DNS
        (atomic).
      - Goto 3.

   7. Sort DNS_RECORDS in descending order according to the last signed line
      number (signed head version 2 or higher).

   8. Get next DNS entry from DNS_RECORDS. If no such entry exists:
      If SKIP_BUILD is true exit with success, otherwise exit with error.

   9. Set SKIP_BUILD to SKIP_BUILD.DNS

  10. Query all TXT records from _codechain-url.DNS and save it as URLs.

  11. If not SKIP_BUILD, validate signed head from TXT (also see ssot package)
      and store HEAD:

      - pubKey from TXT must be the same as pubKey or pubKeyRotate from
        DISK.DNS, if the signed head from DISK.DNS is not expired.
      - The counter from TXT must be larger than the counter from DISK.DNS.
      - The signed head must be valid (as defined by validFrom and validTo).

      If the validation fails, report error. Goto 7.

  12. If not SKIP_BUILD and if signed head from TXT record is the same as the
      one from DISK.DNS, set SKIP_BUILD to true.

  13. If SKIP_BUILD, check if HEAD is contained in
      ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain.
      If not, set SKIP_BUILD to false.
      This can happend if we checked for updates.

  14. Select next URL from URLs. If no such URL exists: Goto 3.

  15. If not SKIP_BUILD, download distribution file from URL/HEAD.tar.gz and
      save it to ~/.config/secpkg/pkgs/NAME/dists
      If it fails: Goto 14.

  16. If not SKIP_BUILD, apply ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz
      to ~/.config/secpkg/pkgs/NAME/src with `codechain apply
      -f ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz -head HEAD`.
      If it fails: Goto 14.

  17. If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
      contains any .secpkg files, ensure these secure dependencies are
      installed and up-to-date. If at least one dependency was updated, set
      SKIP_BUILD to false.

  18. If not SKIP_BUILD, call `make prefix=~/.config/secpkg/local uninstall` in
      ~/.config/secpkg/pkgs/NAME/installed

  19. If not SKIP_BUILD, `rm -rf ~/.config/secpkg/pkgs/NAME/build`

  20. If not SKIP_BUILD,
      `cp -r ~/.config/secpkg/pkgs/NAME/src ~/.config/secpkg/pkgs/NAME/build`

  21. If not SKIP_BUILD, call `make prefix=~/.config/secpkg/local` in
      ~/.config/secpkg/pkgs/NAME/build

  22. If not SKIP_BUILD, call `make prefix= ~/.config/secpkg/local install` in
      ~/.config/secpkg/pkgs/NAME/build

  23. If not SKIP_BUILD,
      `mv ~/.config/secpkg/pkgs/NAME/build ~/.config/secpkg/pkgs/NAME/installed`

  24. If not SKIP_BUILD and the file
      ~/.config/secpkg/pkgs/NAME/installed/.secpkg exists,
      `cp -f ~/.config/secpkg/pkgs/NAME/installed/.secpkg
             ~/.config/secpkg/pkgs/NAME/.secpkg`

  25. If SKIP_BUILD: Goto 8.

  26. The software has been successfully updated.

CheckUpdate specification

Checking if a software package with NAME needs an update works as follows:

   1. Make sure the project with NAME has been installed before.
      That is, the directory ~/.config/secpkg/pkgs/NAME exists.
      Set NEEDS_UPDATE to false.

   2. Load .secpkg file from ~/.config/secpkg/pkgs/NAME/.secpkg

   3. Get next DNS entry from DNS_RECORDS. Set SKIP_CHECK.DNS to false.
      If no such entry exists: Goto 6.

   4. Load signed head from ~/.config/secpkg/pkgs/NAME/signed_head.DNS
      (as DISK.DNS)

   5. Query TXT record from _codechain-head.DNS, if it is the same as DISK.DNS,
      set SKIP_CHECK.DNS to true.

   6. Update signed head:

      - `cp -f ~/.config/secpkg/pkgs/NAME/signed_head.DNS
               ~/.config/secpkg/pkgs/NAME/previous_signed_head.DNS`
      - Save new signed head to ~/.config/secpkg/pkgs/NAME/signed_head.DNS
        (atomic).
      - Goto 3.

   7. Sort DNS_RECORDS in descending order according to the last signed line
      number (signed head version 2 or higher).

   8. Get next DNS entry from DNS_RECORDS. If no such entry exists:
      If SKIP_CHECK is true return NEEDS_UPDATE, otherwise exit with error.

   9. Set SKIP_CHECK to SKIP_CHECK.DNS

  10. If not SKIP_CHECK, validate signed head from TXT (also see ssot package)
      and store HEAD:

      - pubKey from TXT must be the same as pubKey or pubKeyRotate from
        DISK.DNS, if the signed head from DISK.DNS is not expired.
      - The counter from TXT must be larger than the counter from DISK.DNS.
      - The signed head must be valid (as defined by validFrom and validTo).

      If the validation fails, report error. Goto 8.

  11. If not SKIP_CHECK and if signed head from TXT record not the same as the
      one from DISK.DNS, set SKIP_CHECK and NEEDS_UPDATE to true.

  12. If not NEEDS_UPDATE, check if HEAD is contained in
      ~/.config/secpkg/pkgs/NAME/src/.codchain/hashchain.
      If not, set NEEDS_UPDATE to true.

  13. If NEEDS_UPDATE is false, check if the directory
      ~/.config/secpkg/pkgs/NAME/src/.secdep exists and contains any .secpkg
      files, ensure these secure dependencies are installed and up-to-date. If
      at least one dependency needs an update, set NEEDS_UPDATE to true.

  14. If SKIP_CHECK: Goto 8.

  15. Return NEEDS_UPDATE.
*/
package secpkg
