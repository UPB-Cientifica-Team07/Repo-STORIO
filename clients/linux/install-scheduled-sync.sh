#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." &&
  pwd
)"

BIN_SOURCE="$ROOT_DIR/clients/linux/bin/sync-client"
BIN_TARGET="$HOME/.local/bin/upb-sync-client"

SYSTEMD_SOURCE="$ROOT_DIR/clients/linux/systemd"
SYSTEMD_TARGET="$HOME/.config/systemd/user"

CONFIG_DIR="$HOME/.config/upb-cientifica"
CONFIG_FILE="$CONFIG_DIR/sync.env"

mkdir -p \
  "$HOME/.local/bin" \
  "$SYSTEMD_TARGET" \
  "$CONFIG_DIR" \
  "$HOME/UPB-Cientifica"

install \
  -m 0700 \
  "$BIN_SOURCE" \
  "$BIN_TARGET"

install \
  -m 0644 \
  "$SYSTEMD_SOURCE/upb-sync.service" \
  "$SYSTEMD_TARGET/upb-sync.service"

install \
  -m 0644 \
  "$SYSTEMD_SOURCE/upb-sync.timer" \
  "$SYSTEMD_TARGET/upb-sync.timer"

if [ ! -f "$CONFIG_FILE" ]; then

  cat > "$CONFIG_FILE" <<EOF_CONFIG
SYNC_SERVER=localhost:50055
SYNC_AUTH_URL=https://localhost:8081
SYNC_USERNAME=CAMBIAR_LOCALMENTE
SYNC_PASSWORD=CAMBIAR_LOCALMENTE
SYNC_DEVICE_ID=$(hostname)-linux-scheduled
SYNC_DIRECTORY=$HOME/UPB-Cientifica
SYNC_TLS_CA_FILE=$ROOT_DIR/security/pki/upb_dev_ca.crt
AUTH_TLS_CA_FILE=$ROOT_DIR/security/pki/upb_dev_ca.crt
EOF_CONFIG

fi

chmod 0600 "$CONFIG_FILE"

systemctl --user daemon-reload

echo
echo "Instalación completada."
echo "Config local:"
echo "$CONFIG_FILE"
echo
echo "El archivo contiene secretos y NO debe versionarse."
