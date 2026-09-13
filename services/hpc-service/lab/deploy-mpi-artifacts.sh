#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

: "${HPC_MPI_HOSTFILE:?Falta HPC_MPI_HOSTFILE}"

SSH_USER="${HPC_SSH_USER:-$USER}"
REMOTE_DIR="${HPC_REMOTE_MPI_DIR:-/tmp/upb-hpc-mpi}"
LOCAL_DIR="$ROOT_DIR/services/hpc-service/mpi"

echo "===== DEPLOY MPI ARTIFACTS ====="

while read -r HOST REST; do

  [ -n "$HOST" ] || continue

  case "$HOST" in
    \#*)
      continue
      ;;
  esac

  if [ "$HOST" = "127.0.0.1" ] ||
     [ "$HOST" = "localhost" ]; then

    mkdir -p "$REMOTE_DIR"

    cp \
      "$LOCAL_DIR/hello_mpi" \
      "$LOCAL_DIR/sleep_mpi" \
      "$REMOTE_DIR/"

    chmod +x \
      "$REMOTE_DIR/hello_mpi" \
      "$REMOTE_DIR/sleep_mpi"

    echo "PASS: local $HOST"

    continue
  fi

  echo "Deploying to $HOST"

  ssh \
    -o BatchMode=yes \
    -o ConnectTimeout=5 \
    "${SSH_USER}@${HOST}" \
    "mkdir -p '$REMOTE_DIR'"

  scp \
    -q \
    "$LOCAL_DIR/hello_mpi" \
    "$LOCAL_DIR/sleep_mpi" \
    "${SSH_USER}@${HOST}:${REMOTE_DIR}/"

  ssh \
    -o BatchMode=yes \
    -o ConnectTimeout=5 \
    "${SSH_USER}@${HOST}" \
    "chmod +x '$REMOTE_DIR/hello_mpi' '$REMOTE_DIR/sleep_mpi'"

  echo "PASS: $HOST"

done < "$HPC_MPI_HOSTFILE"

echo "PASS: MPI artifacts desplegados"
