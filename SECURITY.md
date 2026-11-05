# Security Policy

## Reporting a Vulnerability

Do not open public issues for security vulnerabilities. Report privately via GitHub Security Advisories or email the maintainer.

## Supported Versions

Only the latest release is supported.

## Cryptography

Orca uses Argon2id (key derivation) and XChaCha20-Poly1305 (AEAD). No custom crypto. Parameters are stored in the authenticated vault header and enforced on load.
