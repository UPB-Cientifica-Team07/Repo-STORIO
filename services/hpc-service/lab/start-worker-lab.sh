#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
    cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
    pwd
)"

: "${COORDINATOR_IP:?Falta COORDINATOR_IP}"
: "${WORKER_IP:?Falta WORKER_IP}"
: "${NODE_ID:?Falta NODE_ID}"

COORDINATOR_PORT="${COORDINATOR_PORT:-1100}"

LOGICAL_HOSTNAME="${LOGICAL_HOSTNAME:-$(hostname)}"

CPU_CORES="${CPU_CORES:-$(nproc)}"

WORKER_RMI_PORT="${WORKER_RMI_PORT:-12001}"

export HPC_WORKER_BIND_ADDRESS="${HPC_WORKER_BIND_ADDRESS:-$WORKER_IP}"

export HPC_WORKER_ADVERTISE_ADDRESS="${HPC_WORKER_ADVERTISE_ADDRESS:-$WORKER_IP}"

export HPC_MPI_DIR="${HPC_MPI_DIR:-$ROOT_DIR/services/hpc-service/mpi}"

if [ -n "${HPC_MPI_HOSTFILE:-}" ]; then
    export HPC_MPI_HOSTFILE
fi

if [ -n "${HPC_MPI_MAX_PROCESSES:-}" ]; then
    export HPC_MPI_MAX_PROCESSES
fi

LOG_DIR="${HPC_LAB_LOG_DIR:-/tmp/upb-hpc-lab}"

mkdir -p "$LOG_DIR"

LOG_FILE="$LOG_DIR/${NODE_ID}.log"
PID_FILE="$LOG_DIR/${NODE_ID}.pid"

echo "===== START HPC WORKER ====="
echo "node=$NODE_ID"
echo "worker_ip=$WORKER_IP"
echo "coordinator=${COORDINATOR_IP}:${COORDINATOR_PORT}"
echo "logical_hostname=$LOGICAL_HOSTNAME"
echo "cpu=$CPU_CORES"
echo "rmi_port=$WORKER_RMI_PORT"

if [ -n "${HPC_MPI_HOSTFILE:-}" ]; then
    echo "mpi_hostfile=$HPC_MPI_HOSTFILE"
fi

nohup bash \
    "$ROOT_DIR/services/hpc-service/scripts/start-worker.sh" \
    "$COORDINATOR_IP" \
    "$COORDINATOR_PORT" \
    "$NODE_ID" \
    "$LOGICAL_HOSTNAME" \
    "$CPU_CORES" \
    "$WORKER_RMI_PORT" \
    > "$LOG_FILE" \
    2>&1 &

PID=$!

echo "$PID" > "$PID_FILE"

sleep 3

if ! kill -0 "$PID" 2>/dev/null; then
    echo "FAIL: Worker termino prematuramente"
    cat "$LOG_FILE"
    exit 1
fi

if ! nc -z "$WORKER_IP" "$WORKER_RMI_PORT"; then
    echo "FAIL: Worker RMI no escucha en ${WORKER_IP}:${WORKER_RMI_PORT}"
    cat "$LOG_FILE"
    exit 1
fi

if ! grep -q 'Estado: REGISTERED' "$LOG_FILE"; then
    echo "FAIL: Worker no quedo REGISTERED"
    cat "$LOG_FILE"
    exit 1
fi

echo "PASS: Worker $NODE_ID PID=$PID"
tail -n 35 "$LOG_FILE"
