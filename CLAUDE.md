# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`digilol-cert-pushpuller` is a Go-based tool for encrypting and syncing TLS certificates via S3 with automatic renewal support using LEGO. It operates in two modes:

- **Push mode** (server): Renews certificates via LEGO, encrypts them, and uploads to S3
- **Pull mode** (client): Downloads and decrypts certificates from S3

## Build Commands

```bash
# Build binary
go build -trimpath -ldflags="-s -w"

# Run tests
go test ./...

# Run specific package tests
go test ./internal/config
go test ./internal/crypto
go test ./internal/util

# Build release packages (requires goreleaser)
goreleaser release --snapshot --clean
```

## Running

```bash
# Push certificates (server mode)
./digilol-cert-pushpuller push --config /path/to/push.toml

# Pull certificates (client mode)
./digilol-cert-pushpuller pull --config /path/to/pull.toml
```

## Architecture

### Core Components

1. **main.go**: Entry point that parses command (`push`/`pull`) and config file, handles daemon mode with scheduling and signal handling

2. **push.go**: Server-side logic

   - Executes LEGO renewal commands (push.go:40-46)
   - Compares local certificate SHA256 hashes with `.hashes.json` from S3 (push.go:54-73)
   - Encrypts changed certificates per-domain using keys from `key_dir` (push.go:106-146)
   - Uploads encrypted `.enc` files to S3 and updates `.hashes.json` (push.go:148-179)
   - Runs reload command (e.g., `systemctl reload nginx`) (push.go:182-187)

3. **pull.go**: Client-side logic
   - Downloads `.hashes.json` from S3 (pull.go:44-63)
   - Lists all `.enc` files in S3 bucket (pull.go:76-94)
   - For each file: checks if local hash matches S3 hash to skip unchanged files (pull.go:128-139)
   - Downloads and decrypts only changed files using matching `.key` files (pull.go:142-167)
   - Runs reload command (pull.go:171-176)

### Internal Packages

- **internal/config**: TOML configuration parsing and encryption key management

  - `LoadPush()`/`LoadPull()`: Parse TOML configs
  - `GetOrCreateKey()`: Per-certificate key management (generates 32-byte keys, stored as base64 in `.key` files with 0600 permissions)
  - Key files named `{certname}.key` (e.g., `_.example.com.key`)

- **internal/crypto**: ChaCha20-Poly1305 encryption using `github.com/minio/sio`

  - `EncryptData()`: Encrypts certificate data
  - `DecryptData()`: Decrypts certificate data

- **internal/s3util**: AWS S3 client creation

  - `NewClient()`: Creates S3 client with custom endpoint support for S3-compatible services

- **internal/util**: Shell command execution
  - `RunCommandWithEnv()`: Executes LEGO commands and reload commands with custom environment variables

### Security Model

**Per-certificate encryption**: Each certificate domain (e.g., `_.example.com`) has its own 256-bit encryption key. This allows selective access control - clients can only decrypt certificates for which they possess the corresponding `.key` file. The server (push) generates keys automatically if they don't exist. Clients (pull) must have the key files distributed separately to decrypt certificates.

### Configuration

- See `push.example.toml` and `pull.example.toml` for complete examples
- Default config location: `/etc/digilol-cert-pushpuller/{push,pull}.toml`
- Supports S3-compatible services via `endpoint` and `force_path_style` options
- Daemon mode available for systems without systemd timers (e.g., Alpine Linux)

### Deployment

Uses GoReleaser (`.goreleaser.yaml`) to build:

- Binaries for Linux/Darwin/FreeBSD (amd64, arm64)
- DEB packages (Debian/Ubuntu) with systemd units
- RPM packages (RHEL/Fedora) with systemd units
- APK packages (Alpine) with OpenRC scripts

Systemd timers and OpenRC scripts are in `packaging/` directory.
