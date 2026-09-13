#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

STAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT_DIR="$ROOT_DIR/reports/hpc/lab-$STAMP"

mkdir -p "$OUTPUT_DIR"

echo "===== HPC EVIDENCE ====="

{
  echo "timestamp=$(date -Iseconds)"
  echo "hostname=$(hostname)"
  echo "git_commit=$(git -C "$ROOT_DIR" rev-parse HEAD)"

  echo
  echo "===== NETWORK ====="
  ip -brief address

  echo
  echo "===== JAVA ====="
  java -version 2>&1

  echo
  echo "===== MPI ====="
  mpirun --version 2>&1 || true

} > "$OUTPUT_DIR/local-system.txt"

if [ -n "${HPC_MPI_HOSTFILE:-}" ] &&
   [ -r "$HPC_MPI_HOSTFILE" ]; then

  cp \
    "$HPC_MPI_HOSTFILE" \
    "$OUTPUT_DIR/mpi-hostfile.txt"

fi

if [ -f /tmp/upb-hpc-direct-mpi.log ]; then

  cp \
    /tmp/upb-hpc-direct-mpi.log \
    "$OUTPUT_DIR/direct-mpi.log"

fi

if [ -f /tmp/upb-hpc-direct-hosts.txt ]; then

  cp \
    /tmp/upb-hpc-direct-hosts.txt \
    "$OUTPUT_DIR/direct-mpi-hosts.txt"

fi

LOG_DIR="${HPC_LAB_LOG_DIR:-/tmp/upb-hpc-lab}"

if [ -d "$LOG_DIR" ]; then

  cp -a \
    "$LOG_DIR" \
    "$OUTPUT_DIR/runtime-logs"

fi

COORDINATOR_IP="${COORDINATOR_IP:-127.0.0.1}"
COORDINATOR_PORT="${COORDINATOR_PORT:-1100}"

if nc -z \
  "$COORDINATOR_IP" \
  "$COORDINATOR_PORT" \
  2>/dev/null
then

  "$ROOT_DIR/services/hpc-service/lab/health-check.sh" \
    > "$OUTPUT_DIR/health.txt" \
    2>&1 \
    || true

fi

echo "output=$OUTPUT_DIR"

find "$OUTPUT_DIR" \
  -type f \
  -print \
  | sort

echo "PASS: evidencia recopilada"
