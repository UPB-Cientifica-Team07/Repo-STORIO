#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

HPC_DIR="$ROOT_DIR/services/hpc-service"
JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"

COORDINATOR_HOST="${1:-127.0.0.1}"
COORDINATOR_PORT="${2:-1100}"
NODE_ID="${3:-node-01}"
LOGICAL_HOSTNAME="${4:-$NODE_ID}"
CPU_CORES="${5:-4}"

cd "$ROOT_DIR"

exec java \
  -cp "$HPC_DIR/bin:$JDBC_JAR" \
  hpc.node.WorkerNode \
  "$COORDINATOR_HOST" \
  "$COORDINATOR_PORT" \
  "$NODE_ID" \
  "$LOGICAL_HOSTNAME" \
  "$CPU_CORES"
