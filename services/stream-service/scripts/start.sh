#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
    cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
    pwd
)"

STREAM_BIND_ADDRESS="${STREAM_BIND_ADDRESS:-127.0.0.1}"

STREAM_PORT="${STREAM_PORT:-8082}"

exec php \
    -d expose_php=0 \
    -S "${STREAM_BIND_ADDRESS}:${STREAM_PORT}" \
    -t "${ROOT_DIR}/services/stream-service/public" \
    "${ROOT_DIR}/services/stream-service/public/router.php"
