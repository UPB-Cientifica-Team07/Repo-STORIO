#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." &&
  pwd
)"

OUTPUT_DIR="$ROOT_DIR/clients/linux/bin"
OUTPUT_FILE="$OUTPUT_DIR/sync-client"

mkdir -p "$OUTPUT_DIR"

cd "$ROOT_DIR/services/sync-service"

echo "Compilando File Sync Client para Linux..."

CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64 \
go build \
  -trimpath \
  -o "$OUTPUT_FILE" \
  ./cmd/watcher

echo
echo "Cliente generado:"
file "$OUTPUT_FILE"
ls -lh "$OUTPUT_FILE"
