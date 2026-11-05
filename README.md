# Orca

Personal password manager. Go core + (planned) Flutter mobile + Syncthing sync.

## Status

Phase 1: Go core + CLI + vault (Argon2id + XChaCha20-Poly1305). In progress.

## Build

```
go build ./...
go test ./...
```

## Usage

```
export ORCA_VAULT=vault.orca
orca init
orca add
orca list
orca get <id>
orca gen
```

## License

Apache-2.0
