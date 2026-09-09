#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -lt 1 ]; then
    echo "Uso:"
    echo "$0 <backup.tar.gz.gpg> [directorio-restauracion]"
    exit 1
fi

BACKUP_FILE="$(realpath "$1")"

RESTORE_DIR="${2:-$(pwd)/restored-backup}"

RESTORE_DIR="$(realpath -m "$RESTORE_DIR")"

GPG_HOME="${BACKUP_GNUPGHOME:-$HOME/.local/share/upb-cientifica/gpg}"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "ERROR: no existe:"
    echo "$BACKUP_FILE"
    exit 1
fi

if [ ! -d "$GPG_HOME" ]; then
    echo "ERROR: no existe GNUPGHOME:"
    echo "$GPG_HOME"
    exit 1
fi

BASENAME="$(basename "$BACKUP_FILE" .tar.gz.gpg)"
BACKUP_DIR="$(dirname "$BACKUP_FILE")"
MANIFEST="$BACKUP_DIR/${BASENAME}.sha256"

TMP_DIR="$(mktemp -d)"
ARCHIVE="$TMP_DIR/${BASENAME}.tar.gz"

cleanup() {
    rm -rf "$TMP_DIR"
}

trap cleanup EXIT

echo "===== RESTORE UPB-CIENTÍFICA ====="
echo "Backup: $BACKUP_FILE"
echo "Destino: $RESTORE_DIR"

echo
echo "[1/4] Descifrando..."

GNUPGHOME="$GPG_HOME" \
gpg \
  --output "$ARCHIVE" \
  --decrypt "$BACKUP_FILE"

echo "[2/4] Verificando SHA-256..."

if [ ! -f "$MANIFEST" ]; then
    echo "ERROR: no existe manifiesto SHA-256:"
    echo "$MANIFEST"
    exit 1
fi

EXPECTED="$(
    awk '{print $1}' "$MANIFEST"
)"

ACTUAL="$(
    sha256sum "$ARCHIVE" \
    | awk '{print $1}'
)"

if [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "ERROR: SHA-256 no coincide"
    echo "Esperado: $EXPECTED"
    echo "Actual:   $ACTUAL"
    exit 1
fi

echo "OK: integridad verificada"

echo "[3/4] Preparando destino..."

mkdir -p "$RESTORE_DIR"

echo "[4/4] Restaurando..."

tar \
  -C "$RESTORE_DIR" \
  -xzf "$ARCHIVE"

echo
echo "===== RESTORE COMPLETADO ====="
echo "Contenido restaurado en:"
echo "$RESTORE_DIR"
