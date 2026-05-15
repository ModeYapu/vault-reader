# Installation

## Quick Start

```bash
# 1. Extract
tar xzf vault-reader-linux-amd64.tar.gz

# 2. Deploy (auto-builds binary, creates systemd service, starts)
sudo bash deploy.sh /opt/obsidian-vault

# 3. Open
open http://localhost:3000
```

## Options

```bash
# Custom vault directory
sudo bash deploy.sh /data/my-vault

# With subpath prefix (e.g. behind reverse proxy at /vault/)
sudo bash deploy.sh /data/my-vault /vault

# Use install-service.sh if you already built the binary
bash scripts/build.sh v0.2.0
sudo bash scripts/install-service.sh /opt/obsidian-vault /opt/vault-reader-data
```

## CLI Flags

```
--vault   (required)  Path to Obsidian Vault
--addr    (default :3000)  Listen address
--data    (default <vault>/.vault-reader-data)  Index database directory
--prefix  (default "")  URL subpath prefix, e.g. /vault
```

Environment variables: `VAULT_DIR`, `ADDR`, `DATA_DIR`, `PREFIX`

## Reverse Proxy

If using `--prefix /vault`, configure Nginx:

```nginx
location /vault/ {
    proxy_pass http://127.0.0.1:3000/vault/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Or Caddy:

```
example.com {
    handle_path /vault/* {
        reverse_proxy localhost:3000
    }
}
```

## Docker

```bash
docker build -t vault-reader .
docker run -d -p 3000:3000 \
  -v /path/to/vault:/vault:ro \
  -v vault-data:/data \
  vault-reader
```

## Management

```bash
systemctl status vault-reader
systemctl restart vault-reader
systemctl stop vault-reader
journalctl -u vault-reader -f
```
