# Security Policy

## Reporting a Vulnerability

Do not open public issues for security vulnerabilities. Report privately via GitHub Security Advisories or email the maintainer.

## Supported Versions

Only the latest release is supported.

## Cryptography

Orca uses Argon2id (key derivation) and XChaCha20-Poly1305 (AEAD). No custom crypto. Parameters are stored in the authenticated vault header and enforced on load.


## Mobile (Flutter) bridge note
Dart `String` is immutable and cannot be zeroed in memory. Master passwords and derived values held in Dart remain until GC. Mitigations: password is passed via stdin only (never argv/env), never logged, and the TextField controller is cleared on unlock success/failure. All crypto stays in the Go core; Flutter handles UI, biometric gating, and storage only. Biometric unlock re-derives from the master password on failure and never caches the derived key.
