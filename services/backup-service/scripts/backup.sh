#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

SOURCE_DIR="${BACKUP_SOURCE_DIR:-$ROOT/services/file-service/data/shared-storage}"
BACKUP_DIR="${BACKUP_OUTPUT_DIR:-$ROOT/services/backup-service/backups}"
PUBLIC_KEY="${BACKUP_PUBLIC_KEY:-$ROOT/services/backup-service/keys/upb-backup-public.asc}"

TIMESTAMP="$(date '+%Y%m%d-%H%M%S')"
BASENAME="upb-backup-${TIMESTAMP}"

TMP_DIR="$(mktemp -d)"
ARCHIVE="$TMP_DIR/${BASENAME}.tar.gz"
HASH_FILE="$TMP_DIR/${BASENAME}.sha256"
ENCRYPTED="$BACKUP_DIR/${BASENAME}.tar.gz.gpg"
MANIFEST="$BACKUP_DIR/${BASENAME}.sha256"

cleanup() {
    rm -rf "$TMP_DIR"
}

trap cleanup EXIT

if [ ! -d "$SOURCE_DIR" ]; then
    echo "ERROR: no existe el directorio fuente:"
    echo "$SOURCE_DIR"
    exit 1
fi

if [ ! -s "$PUBLIC_KEY" ]; then
    echo "ERROR: no existe la clave pública GPG:"
    echo "$PUBLIC_KEY"
    exit 1
fi

mkdir -p "$BACKUP_DIR"

GPG_HOME="$(mktemp -d)"
chmod 700 "$GPG_HOME"

cleanup_gpg() {
    rm -rf "$GPG_HOME"
}

trap 'cleanup; cleanup_gpg' EXIT

echo "===== BACKUP UPB-CIENTÍFICA ====="
echo "Fuente: $SOURCE_DIR"
echo "Destino: $ENCRYPTED"

echo
echo "[1/4] Empaquetando..."

tar \
  -C "$(dirname "$SOURCE_DIR")" \
  -czf "$ARCHIVE" \
  "$(basename "$SOURCE_DIR")"

echo "[2/4] SHA-256..."

(
    cd "$TMP_DIR"
    sha256sum "$(basename "$ARCHIVE")" \
      > "$(basename "$HASH_FILE")"
)

cp "$HASH_FILE" "$MANIFEST"

echo "[3/4] Importando clave pública..."

GNUPGHOME="$GPG_HOME" \
gpg \
  --batch \
  --quiet \
  --import "$PUBLIC_KEY"

FINGERPRINT="$(
    GNUPGHOME="$GPG_HOME" \
    gpg \
      --batch \
      --with-colons \
      --fingerprint \
      "backup@upb-cientifica.local" \
    | awk -F: '$1 == "fpr" {print $10; exit}'
)"

if [ -z "$FINGERPRINT" ]; then
    echo "ERROR: no se pudo obtener fingerprint de la clave pública"
    exit 1
fi

echo "[4/4] Cifrando con GPG..."

GNUPGHOME="$GPG_HOME" \
gpg \
  --batch \
  --yes \
  --trust-model always \
  --recipient "$FINGERPRINT" \
  --output "$ENCRYPTED" \
  --encrypt "$ARCHIVE"

chmod 600 "$ENCRYPTED"
chmod 644 "$MANIFEST"

echo
echo "===== BACKUP COMPLETADO ====="
echo "Archivo cifrado:"
echo "$ENCRYPTED"
echo
echo "Hash del contenido original:"
cat "$MANIFEST"
echo
echo "El archivo .tar.gz sin cifrar fue eliminado con el directorio temporal."
