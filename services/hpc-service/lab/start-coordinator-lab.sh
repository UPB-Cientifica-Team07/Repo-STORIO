#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
    cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
    pwd
)"

: "${COORDINATOR_IP:?Falta COORDINATOR_IP}"
: "${HPC_DB_PASSWORD:?Falta HPC_DB_PASSWORD}"

HPC_RMI_PORT="${HPC_RMI_PORT:-1100}"
HPC_COORDINATOR_RMI_PORT="${HPC_COORDINATOR_RMI_PORT:-1101}"

export HPC_BIND_ADDRESS="${HPC_BIND_ADDRESS:-$COORDINATOR_IP}"
export HPC_ADVERTISE_ADDRESS="${HPC_ADVERTISE_ADDRESS:-$COORDINATOR_IP}"

export HPC_RMI_PORT
export HPC_COORDINATOR_RMI_PORT

export HPC_AUTH_SERVICE="${HPC_AUTH_SERVICE:-https://127.0.0.1:8081}"

LOG_DIR="${HPC_LAB_LOG_DIR:-/tmp/upb-hpc-lab}"

mkdir -p "$LOG_DIR"

echo "===== START HPC COORDINATOR ====="
echo "ip=$COORDINATOR_IP"
echo "registry_port=$HPC_RMI_PORT"
echo "object_port=$HPC_COORDINATOR_RMI_PORT"
echo "log=$LOG_DIR/coordinator.log"

nohup bash \
    "$ROOT_DIR/services/hpc-service/scripts/start-coordinator.sh" \
    > "$LOG_DIR/coordinator.log" \
    2>&1 &

PID=$!

echo "$PID" \
    > "$LOG_DIR/coordinator.pid"

sleep 2

if ! kill -0 "$PID" 2>/dev/null; then
    echo "FAIL: Coordinator termino prematuramente"
    cat "$LOG_DIR/coordinator.log"
    exit 1
fi

if ! nc -z "$COORDINATOR_IP" "$HPC_RMI_PORT"; then
    echo "FAIL: RMI registry no escucha en ${COORDINATOR_IP}:${HPC_RMI_PORT}"
    cat "$LOG_DIR/coordinator.log"
    exit 1
fi

echo "PASS: Coordinator PID=$PID"
tail -n 30 "$LOG_DIR/coordinator.log"
