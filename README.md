# digilol-cert-pushpuller

Tool for encrypting and syncing certificates via S3 with automatic renewal support using [LEGO](https://go-acme.github.io/lego/usage/cli/renew-a-certificate/index.html).

## Quick Start

**1. Install the package:**

```bash
# Detect architecture
ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

# Debian/Ubuntu
PACKAGE="$(curl -s "https://api.github.com/repos/digilolnet/digilol-cert-pushpuller/releases/latest" | grep -oP "\"browser_download_url\": \"\K[^\"]*linux_${ARCH}\.deb")"
curl -LO "$PACKAGE"
sudo dpkg -i "$(basename "$PACKAGE")"

# RHEL/Fedora/CentOS
PACKAGE="$(curl -s "https://api.github.com/repos/digilolnet/digilol-cert-pushpuller/releases/latest" | grep -oP "\"browser_download_url\": \"\K[^\"]*linux_${ARCH}\.rpm")"
curl -LO "$PACKAGE"
sudo rpm -i "$(basename "$PACKAGE")"

# Alpine Linux
PACKAGE="$(curl -s "https://api.github.com/repos/digilolnet/digilol-cert-pushpuller/releases/latest" | grep -oP "\"browser_download_url\": \"\K[^\"]*linux_${ARCH}\.apk")"
curl -LO "$PACKAGE"
sudo apk add --allow-untrusted "$(basename "$PACKAGE")"

# Or download binary archive for any OS
PACKAGE="$(curl -s "https://api.github.com/repos/digilolnet/digilol-cert-pushpuller/releases/latest" | grep -oP "\"browser_download_url\": \"\K[^\"]*linux_${ARCH}\.tar\.gz")"
curl -LO "$PACKAGE"
tar xzf "$(basename "$PACKAGE")" digilol-cert-pushpuller
sudo mv digilol-cert-pushpuller /usr/local/bin/
```

**2. Configure:**

For push (server), edit `/etc/digilol-cert-pushpuller/push.toml`:

```toml
key_dir = "/var/lib/digilol-cert-pushpuller/keys"
cert_dir = ".lego/certificates"
reload_cmd = "systemctl reload nginx"

[daemon]
enabled = false
interval_secs = 86400
jitter_secs = 3600

[s3]
bucket = "my-certificates-bucket"
endpoint = "https://s3.example.com"
region = "us-east-1"
prefix = ""
force_path_style = true
access_key = "your-s3-access-key"
secret_key = "your-s3-secret-key"

[[lego_commands]]
command = "lego -d '*.example.com' -d example.com -a -m admin@example.com --dns cloudflare renew"
[lego_commands.env]
CF_DNS_API_TOKEN = "your-cloudflare-token"

[[lego_commands]]
command = "lego -d '*.example.net' -d example.net -s https://acme.zerossl.com/v2/DV90 -a -m admin@example.net --eab --kid your-eab-kid --hmac your-eab-hmac --dns bunny renew"
[lego_commands.env]
BUNNY_API_KEY = "your-bunny-api-key"
```

For pull (client), edit `/etc/digilol-cert-pushpuller/pull.toml`:

```toml
key_dir = "/var/lib/digilol-cert-pushpuller/keys"
cert_dir = "/var/lib/digilol-cert-pushpuller/certificates"
reload_cmd = "systemctl reload nginx"

[daemon]
enabled = false
interval_secs = 300
jitter_secs = 0

[s3]
bucket = "my-certificates-bucket"
endpoint = "https://s3.example.com"
region = "us-east-1"
prefix = ""
force_path_style = true
access_key = "your-s3-access-key"
secret_key = "your-s3-secret-key"
```

**3. Enable services:**

```bash
# Debian/Ubuntu/RHEL/Fedora (systemd)
sudo systemctl enable --now digilol-cert-pushpuller-push.timer # For server (push)
sudo systemctl enable --now digilol-cert-pushpuller-pull.timer # For client (pull)

# Alpine Linux (OpenRC)
sudo rc-update add digilol-cert-pushpuller-push default
sudo rc-update add digilol-cert-pushpuller-pull default
sudo rc-service digilol-cert-pushpuller-push start
sudo rc-service digilol-cert-pushpuller-pull start
```

Done! The server will renew and push certificates daily (with 1h random delay), clients will pull every 5 minutes.

**Note:** Supports Linux, macOS, and FreeBSD. Packages available for Debian, Ubuntu, RHEL, Fedora, CentOS, and Alpine Linux.

**Daemon Mode:** On Alpine Linux or systems without systemd timers, you can enable daemon mode by setting `daemon.enabled = true` in the config file. The tool will run continuously with built-in scheduling instead of relying on external cron/timer systems.

## Configuration

See [push.example.toml](push.example.toml) and [pull.example.toml](pull.example.toml) for complete examples.

**Push config fields:**

- `key_dir`: Directory for encryption keys
- `cert_dir`: Directory containing certificates to push
- `lego_commands`: Array of lego renewal commands (optional)
- `reload_cmd`: Command to run after push (optional)
- `daemon.enabled`: Enable daemon mode (default: false)
- `daemon.interval_secs`: Seconds between runs (default: 86400 for push)
- `daemon.jitter_secs`: Random delay in seconds (default: 3600 for push)
- `s3.bucket`: S3 bucket name
- `s3.region`: S3 region
- `s3.access_key`: S3 access key
- `s3.secret_key`: S3 secret key
- `s3.endpoint`: S3 endpoint URL
- `s3.prefix`: S3 key prefix/folder (optional)
- `s3.force_path_style`: Use path-style URLs (required for most S3-compatible services)

**Pull config fields:**

- `key_dir`: Directory for encryption keys
- `cert_dir`: Directory to store pulled certificates
- `reload_cmd`: Command to run after pull (optional)
- `daemon.enabled`: Enable daemon mode (default: false)
- `daemon.interval_secs`: Seconds between runs (default: 300 for pull)
- `daemon.jitter_secs`: Random delay in seconds (default: 0 for pull)
- `s3.bucket`: S3 bucket name
- `s3.region`: S3 region
- `s3.access_key`: S3 access key
- `s3.secret_key`: S3 secret key
- `s3.endpoint`: S3 endpoint URL
- `s3.prefix`: S3 key prefix/folder (optional)
- `s3.force_path_style`: Use path-style URLs (required for most S3-compatible services)

## How It Works

**Push (server):**

1. Runs lego commands to renew certificates
2. Calculates SHA256 checksum of each certificate file
3. Compares with `.hashes.json` in S3 to skip unchanged files
4. Encrypts changed certificates with unique keys
5. Uploads encrypted `.enc` files to S3
6. Updates `.hashes.json` in S3
7. Runs reload command

**Pull (client):**

1. Downloads `.hashes.json` from S3
2. Lists `.enc` files in S3
3. For each file, checks if local file exists and compares SHA256 checksum
4. Skips download if checksum matches (file unchanged)
5. Downloads and decrypts only changed files (and only if local encryption key exists)
6. Runs reload command

**Security:**

- Each certificate domain has a unique 256-bit encryption key (per-certificate encryption allows selective access: clients can only decrypt certificates for which they have the corresponding key file)
- Keys stored as raw 32-byte `.key` files with 0600 permissions
- Encryption: ChaCha20-Poly1305 via `github.com/minio/sio`
- Clients only pull and decrypt certificates for which they have the key files

## Manual Usage

```bash
# Push certificates
digilol-cert-pushpuller push --config /etc/digilol-cert-pushpuller/push.toml

# Pull certificates
digilol-cert-pushpuller pull --config /etc/digilol-cert-pushpuller/pull.toml
```

## Building from Source

```bash
go build -trimpath -ldflags="-s -w"
```

## License

Apache License 2.0 - see [LICENSE.txt](LICENSE.txt)
