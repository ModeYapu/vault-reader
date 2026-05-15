#!/bin/bash
# deploy.sh — vault-reader Linux 部署脚本
# 用法: sudo bash deploy.sh /opt/obsidian-vault

set -euo pipefail

VAULT_DIR="${1:-/opt/obsidian-vault}"
INSTALL_DIR="/opt/vault-reader"
DATA_DIR="/opt/vault-reader-data"
SERVICE_USER="vaultreader"

echo "=== vault-reader 部署 ==="
echo "Vault 目录: ${VAULT_DIR}"
echo "安装目录:   ${INSTALL_DIR}"
echo ""

# 1. 编译 Linux amd64 二进制
echo "[1/5] 编译 Linux 二进制..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o vault-reader-linux ./cmd/vault-reader

# 2. 创建用户（如果不存在）
echo "[2/5] 创建服务用户..."
if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

# 3. 安装文件
echo "[3/5] 安装..."
mkdir -p "${INSTALL_DIR}"
cp vault-reader-linux "${INSTALL_DIR}/vault-reader"
chmod +x "${INSTALL_DIR}/vault-reader"
mkdir -p "${DATA_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}"

# 4. 写入 systemd 服务文件
echo "[4/5] 配置 systemd 服务..."
cat > /etc/systemd/system/vault-reader.service << EOF
[Unit]
Description=Vault Reader - Obsidian Vault Web Reader
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
ExecStart=${INSTALL_DIR}/vault-reader --vault ${VAULT_DIR} --data ${DATA_DIR} --addr :3000
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

# 安全加固
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=${VAULT_DIR}
ReadWritePaths=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF

# 5. 启动服务
echo "[5/5] 启动服务..."
systemctl daemon-reload
systemctl enable vault-reader
systemctl restart vault-reader

echo ""
echo "=== 部署完成 ==="
echo "服务状态: systemctl status vault-reader"
echo "查看日志: journalctl -u vault-reader -f"
echo "访问地址: http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo localhost):3000"
echo ""
echo "常用命令:"
echo "  重启: systemctl restart vault-reader"
echo "  停止: systemctl stop vault-reader"
echo "  日志: journalctl -u vault-reader --since today"
