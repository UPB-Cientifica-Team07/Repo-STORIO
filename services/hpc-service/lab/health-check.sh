#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

COORDINATOR_IP="${COORDINATOR_IP:-127.0.0.1}"
COORDINATOR_PORT="${COORDINATOR_PORT:-1100}"
JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"

echo "===== HPC HEALTH CHECK ====="
echo "coordinator=${COORDINATOR_IP}:${COORDINATOR_PORT}"

if ! nc -z \
  "$COORDINATOR_IP" \
  "$COORDINATOR_PORT"
then
  echo "FAIL: Coordinator no accesible"
  exit 1
fi

echo "PASS: RMI registry accesible"

java \
  -cp "$ROOT_DIR/services/hpc-service/bin:$JDBC_JAR" \
  hpc.client.HealthClient \
  "$COORDINATOR_IP" \
  "$COORDINATOR_PORT"
