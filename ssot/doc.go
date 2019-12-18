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
  zone (if not published automatically). DNSSEC should be enabled.

SignHead specification

To publish an update of a secure package with SSOT do the following:

   1. Parse the .secpkg file in the current working directory.

   2. Make sure the project with NAME has been published before.
      That is, the directory ~/.config/ssotpub/pkgs/NAME exists.

   3. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head.

   4. Get the HEAD from .codechain/hashchain in the current working directory.

   5. If ~/.config/ssotpub/pkgs/NAME/cloudflare.json exits, check the contained
      Cloudflare credentials and switch on automatic publishing of TXT records.

   6. If ROTATE is set, check if ~/.config/ssotput/pkgs/NAME/rotate_to exists.
      If it does, abort. Otherwise write public key to rotate to and rotate time
      (see below) to ~/.config/ssotput/pkgs/NAME/rotate_to.

   7. Create a new signed head with current HEAD, the counter of the previous
      signed head plus 1, and update the saved signed head:

      - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
               ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
      - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).

      If ~/.config/ssotput/pkgs/NAME/rotate_to exists:

      - If rotate time has been reached use pubkey from file as PUBKEY and
        remove ~/.config/ssotput/pkgs/NAME/rotate_to.
      - Otherwise use old PUBKEY and set pubkey from file as PUBKEY_ROTATE.

   8. If the HEAD changed, save the current distribution to:
      ~/.config/secpkg/pkgs/NAME/dists/HEAD.tar.gz (`codechain createdist`).

   9. If the HEAD changed, lookup the download URL and print where to upload
      the distribution file:
      ~/.config/ssotpkg/pkgs/NAME/dists/HEAD.tar.gz

  10. Print DNS TXT record as defined by the .secpkg file and the signed head.
      If TXT records are to be published automatically, publish the TXT record.

  11. If the HEAD changed, update the .secpkg file accordingly.

  Afterwards the administrator manually uploads the distribution HEAD.tar.gz
  to the download URL and publishes the new DNS TXT record in the defined
  zone (if not published automatically). DNSSEC should be enabled.

Refresh specification

To refresh the published head of a secure package with SSOT do the following:

   1. Parse the supplied .secpkg file.

   2. Make sure the project with NAME has been published before.
      That is, the directory ~/.config/ssotpub/pkgs/NAME exists.

   3. Validate the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head.

   4. Make sure the signed head in ~/.config/ssotpub/pkgs/NAME/signed_head
      matches the HEAD in the .secpkg file.

   5. If ~/.config/ssotpub/pkgs/NAME/cloudflare.json exits, check the contained
      Cloudflare credentials and switch on automatic publishing of TXT records.

   6. If ROTATE is set, check if ~/.config/ssotput/pkgs/NAME/rotate_to exists.
      If it does, abort. Otherwise write public key to rotate to and rotate time
      (see below) to ~/.config/ssotput/pkgs/NAME/rotate_to.

   7. Create a new signed head with the same HEAD, the counter of the previous
      signed head plus 1, and update the saved signed head:

      - `cp -f ~/.config/ssotpub/pkgs/NAME/signed_head
               ~/.config/ssotpub/pkgs/NAME/previous_signed_head`
      - Save new signed head to ~/.config/ssotpub/pkgs/NAME/signed_head (atomic).

      If ~/.config/ssotput/pkgs/NAME/rotate_to exists:

      - If rotate time has been reached use pubkey from file as PUBKEY and
        remove ~/.config/ssotput/pkgs/NAME/rotate_to.
      - Otherwise use old PUBKEY and set pubkey from file as PUBKEY_ROTATE.

   8. Print DNS TXT record as defined by the .secpkg file and the signed head.
      If TXT record is to be published automatically, publish the TXT record.

  Afterwards the administrator publishes the new DNS TXT record in the defined
  zone (if not published automatically). DNSSEC should be enabled.

Rotate time calculation

The earliest time a PUBKEY_ROTATE can be used as PUBKEY is when the previous
signed head (without PUBKEY_ROTATE) has expired. This gives clients time to
learn about PUBKEY_ROTATE. To give some extra time we take the time span a
signed head with PUBKEY_ROTATE is valid after the signed head without
PUBKEY_ROTATE has expired and divide it by three. The rotate time is set to the
end of the first third.
*/
package ssot
