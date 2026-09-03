#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

HPC_DIR="$ROOT_DIR/services/hpc-service"
JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"

cd "$ROOT_DIR"

exec java \
  -cp "$HPC_DIR/bin:$JDBC_JAR" \
  hpc.coordinator.CoordinatorServer
