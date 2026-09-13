#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

cd "$ROOT_DIR"

: "${GRID_DB_PASSWORD:=${POSTGRES_PASSWORD:-}}"

if [ -z "$GRID_DB_PASSWORD" ]; then
  echo "ERROR: GRID_DB_PASSWORD o POSTGRES_PASSWORD es obligatorio" >&2
  exit 1
fi

export GRID_DB_PASSWORD

export POSTGRES_HOST="${POSTGRES_HOST:-127.0.0.1}"
export POSTGRES_PORT="${POSTGRES_PORT:-5434}"
export POSTGRES_USER="${POSTGRES_USER:-upb_app}"
export POSTGRES_DB="${POSTGRES_DB:-upb_cientifica}"

export GRID_BIND_ADDRESS="${GRID_BIND_ADDRESS:-127.0.0.1:50056}"
export GRID_AUTH_SERVICE="${GRID_AUTH_SERVICE:-https://127.0.0.1:8081}"

export AUTH_TLS_CA_FILE="${
  AUTH_TLS_CA_FILE:-security/pki/upb_dev_ca.crt
}"

export GRID_WATCHDOG_INTERVAL="${
  GRID_WATCHDOG_INTERVAL:-5s
}"

export GRID_RESOURCE_TIMEOUT="${
  GRID_RESOURCE_TIMEOUT:-15s
}"

exec go run \
  ./services/grid-service/cmd/server
