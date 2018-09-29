/*
Package ssot implements a single source of truth (SSOT) with DNS TXT records.

CreatePkg specification

To create a new secure package for a project developed with Codechain that
should be distributed with a SSOT using DNS TXT records, the following
procedure is defined:

  1. Make sure the project with NAME has not been published before.
     That is, the directory ~/.config/ssotpub/pkgs/NAME does not exist.

  2. Create a new .secpkg file which specifies the following:

     - The NAME of the project.
     - The fully qualified domain name (DNS) where the TXT records can be
       queried.
     - The URL under which the distribution .tar.gz files can be downloaded.
     - The current HEAD of the project's Codechain.

     The .secpkg file is saved to the current working directory, which is
     typically added to the root of the project's repository.

  3. Create the first signed head (see SignHead) for the current project's
     HEAD with a supplied secret key and counter set to 0.

  4. Create the directory ~/.config/ssotpub/pkgs/NAME and save the signed head
     to ~/.config/ssotpub/pkgs/NAME/signed_head

  5. Print DNS TXT record as defined by the .secpkg and the first signed head.

  Afterward the administrator manually publishes the new DNS TXT record in the
  defined zone. DNSSEC should be enabled.

SignHead specification

To publish an update of a secure package with SSOT do the following:

  1. Make sure the project with NAME has been published before.
     That is, the directory ~/.config/ssotpub/pkgs/NAME exists.

  2. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
     and make sure the corresponding secret key is available.

  3. Parse the .secpkg file in the current working directory.

  4. Get the HEAD from .codechain/hashchain in the current working directory.

  5. Create a new signed head with current HEAD, the counter of the previous
     signed head plus 1, and update the saved signed head:

     - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
              ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
     - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).

  6. Print DNS TXT record as defined by the .secpkg and the signed head.

  7. If the HEAD changed, update the .secpkg file accordingly.

  Afterward the administrator manually publishes the updated DNS TXT record in
  the defined zone. DNSSEC should be enabled.

TODO

The following should be specified:

  - Key rotation.
  - Automatic publishing of TXT records. With https://github.com/xenolf/lego?
*/
package ssot
