# Security Policy

## Supported versions

The latest minor release receives security fixes. Older versions: upgrade.

| Version | Supported |
|---|---|
| latest release | yes |
| anything older | no |

## Token handling

- torre targets Torre.ai's **public** API, which needs no credentials. A bearer
  token is optional (only for endpoints that require one).
- When set, the token is stored in the OS keyring (macOS Keychain, Linux Secret
  Service, Windows Credential Manager) under the service name `torre-cli`, keyed
  `<profile>`.
- On hosts without a keyring, the token falls back to an AES-256-GCM encrypted file
  (`credentials.enc`, mode 0600) under the config dir. The key derives from
  `TORRE_KEYRING_PASSWORD` via scrypt when set; otherwise from a host-bound seed,
  which is obfuscation, not a security boundary — set the password on shared hosts.
- The token never appears in `config.yaml`, command output, or `--dry-run` curls
  (redacted unless you pass `--show-token`).
- Cleartext `http://` base URLs are rejected for non-loopback hosts.
- `torre auth logout` removes exactly the active profile's stored token.

## Reporting a vulnerability

Please use GitHub's private vulnerability reporting on this repository
(Security → Report a vulnerability). Do not open a public issue for anything
sensitive. Reports get a response within a week.
