/*
Package ssot implements a single source of truth (SSOT) with DNS TXT records.

Signed head specification

Signed heads have the following fields:

  - PUBKEY (32-byte), the Ed25519 public key of SSOT head signer.
  - PUBKEY_ROTATE (32-byte), Ed25519 pubkey to rotate to, set to 0 if unused.
  - VALID_FROM (8-byte), the signed head is valid from the given Unix time.
  - VALID_TO (8-byte), the signed head is valid to the given Unix time.
  - COUNTER (8-byte), strictly increasing signature counter.
  - HEAD, the Codechain head to sign.
  - SIGNATURE, signature with PUBKEY.

The SIGNATURE is over all previous fields:

  PUBKEY|PUBKEY_ROTATE|VALID_FROM|VALID_TO|COUNTER|HEAD

The signed head is a concatenation of

  PUBKEY|PUBKEY_ROTATE|VALID_FROM|VALID_TO|COUNTER|HEAD|SIGNATURE

encoded in base64 (URL encoding without padding).

All integers (VALID_FROM, VALID_TO, COUNTER) are encoded in network order
(big-endian).

CreatePkg specification

To create a new secure package for a project developed with Codechain that
should be distributed with a SSOT using DNS TXT records, the following
procedure is defined:

  1. Make sure the project with NAME has not been published before.
     That is, the directory ~/.config/ssotpub/pkgs/NAME does not exist.

  2. If TXT records are to be published automatically, check credentials.

  3. Create a new .secpkg file which specifies the following:

     - The NAME of the project.
     - The fully qualified domain name (DNS) where the TXT records can be
       queried.
     - The current HEAD of the project's Codechain.

     The .secpkg file is saved to the current working directory, which is
     typically added to the root of the project's repository.

  4. Create the first signed head (see SignHead) for the current project's
     HEAD with a supplied secret key and counter set to 0.

  5. Create the directory ~/.config/ssotpub/pkgs/NAME/dists
     and save the current distribution to
      ~/.config/ssotpub/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`).

  6. Save the signed head to ~/.config/ssotpub/pkgs/NAME/signed_head

  7. Print the distribution name: ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz

  8. Print DNS TXT records as defined by the .secpkg, the first signed head,
     and the download URL. If TXT records are to be published automatically,
     save credentials and publish the TXT record.

  Afterwards the administrator manually uploads the distribution HEAD.tar.gz
  to the download URL and publishes the new DNS TXT record in the defined
  zone. DNSSEC should be enabled.

SignHead specification

To publish an update of a secure package with SSOT do the following:

  1. Parse the .secpkg file in the current working directory.

  2. Make sure the project with NAME has been published before.
     That is, the directory ~/.config/ssotpub/pkgs/NAME exists.

  3. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
     and make sure the corresponding secret key is available.

  4. Get the HEAD from .codechain/hashchain in the current working directory.

  5. Create a new signed head with current HEAD, the counter of the previous
     signed head plus 1, and update the saved signed head:

     - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
              ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
     - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).

  6. If the HEAD changed, save the current distribution to:
     ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`).

  7. If the HEAD changed, lookup the download URL and print where to upload
     the distribution file:
     ~/.config/ssotpkg/pkgs/NAME/dists/HEAD.tar.gz

  8. Print DNS TXT record as defined by the .secpkg and the signed head.

  9. If the HEAD changed, update the .secpkg file accordingly.

  Afterwards the administrator manually uploads the distribution HEAD.tar.gz
  to the download URL and publishes the new DNS TXT record in the defined
  zone. DNSSEC should be enabled.

TODO

The following should be specified:

  - Key rotation.
  - Automatic publishing of TXT records (with dyn package).
*/
package ssot
