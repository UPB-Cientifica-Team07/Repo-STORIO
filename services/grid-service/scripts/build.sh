#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

cd "$ROOT_DIR/services/grid-service"

mkdir -p bin

go build \
  -o bin/grid-server \
  ./cmd/server

echo "PASS: bin/grid-server"
