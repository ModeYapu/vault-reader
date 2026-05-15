#!/bin/bash
set -e

# Install vault-reader as a systemd service
# Usage: sudo ./scripts/install-service.sh [/opt/obsidian-vault] [/opt/vault-reader-data] [/vault]
# 示例:
#   sudo ./scripts/install-service.sh                          # 默认
#   sudo ./scripts/install-service.sh /data/vault              # 指定 vault
#   sudo ./scripts/install-service.sh /data/vault /data /vault # 指定 vault + data + prefix

VAULT_DIR="${1:-/opt/obsidian-vault}"
DATA_DIR="${2:-/opt/vault-reader-data}"
PREFIX="${3:-}"
BIN_PATH="/usr/local/bin/vault-reader"
SERVICE_USER="vaultreader"

if [ ! -f "bin/vault-reader" ]; then
    echo "Error: bin/vault-reader not found. Run scripts/build.sh first."
    exit 1
fi

echo "Installing vault-reader..."
echo "  Vault:  ${VAULT_DIR}"
echo "  Data:   ${DATA_DIR}"
echo "  Prefix: ${PREFIX:-none (root)}"

# Stop service if running
systemctl stop vault-reader 2>/dev/null || true

# Copy binary
cp bin/vault-reader "${BIN_PATH}"
chmod +x "${BIN_PATH}"

# Create data directory
mkdir -p "${DATA_DIR}"

# Create systemd service with security hardening
cat > /etc/systemd/system/vault-reader.service << EOF
[Unit]
Description=Vault Reader - Obsidian Vault Web Reader
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
ExecStart=${BIN_PATH} --vault ${VAULT_DIR} --data ${DATA_DIR} --addr :3000${PREFIX:+ --prefix ${PREFIX}}
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=${VAULT_DIR}
ReadWritePaths=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF

# Create service user if not exists
if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

# Set permissions
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}"
chown "${SERVICE_USER}:${SERVICE_USER}" "${BIN_PATH}"

# Enable and start
systemctl daemon-reload
systemctl enable vault-reader
systemctl restart vault-reader

echo "Done! Service installed and started."
echo "  Status: systemctl status vault-reader"
echo "  Logs:   journalctl -u vault-reader -f"
echo "  URL:    http://localhost:3000${PREFIX}"
if [ -n "${PREFIX}" ]; then
echo ""
echo "⚠️  子路径 ${PREFIX} 已启用，请配置反向代理转发到 localhost:3000"
fi
