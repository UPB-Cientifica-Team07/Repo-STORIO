#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

cd "$ROOT"

echo "========================================"
echo " GPG BACKUP SERVICE - VERIFICACIÓN"
echo "========================================"

echo
echo "[1/5] GPG"

gpg --version \
  | head -n 1

echo
echo "[2/5] Clave pública"

gpg \
  --show-keys \
  services/backup-service/keys/upb-backup-public.asc \
  >/dev/null

echo "OK: clave pública válida"

echo
echo "[3/5] Scripts"

test -x services/backup-service/scripts/backup.sh
test -x services/backup-service/scripts/restore.sh

echo "OK: scripts ejecutables"

echo
echo "[4/5] Clave privada en repo"

if find services/backup-service \
    -type f \
    \( \
      -iname '*private*' \
      -o -iname '*.sec' \
      -o -iname '*.key' \
    \) \
    | grep -q .
then
    echo "ERROR: se encontró material privado"
    exit 1
fi

echo "OK: sin clave privada"

echo
echo "[5/5] Shell"

bash -n services/backup-service/scripts/backup.sh
bash -n services/backup-service/scripts/restore.sh

echo "OK: sintaxis válida"

echo
echo "========================================"
echo " GPG BACKUP SERVICE: VERIFICACIÓN OK"
echo "========================================"
