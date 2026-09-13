#!/usr/bin/env bash

set -euo pipefail

: "${HPC_MPI_HOSTFILE:?Falta HPC_MPI_HOSTFILE}"

SSH_USER="${HPC_SSH_USER:-$USER}"
MPI_DIR="${HPC_REMOTE_MPI_DIR:-/tmp/upb-hpc-mpi}"

FAIL=0
HOSTS=0

echo "===== SSH + MPI PREFLIGHT ====="

while read -r HOST REST; do

  [ -n "$HOST" ] || continue

  case "$HOST" in
    \#*)
      continue
      ;;
  esac

  HOSTS=$((HOSTS + 1))

  echo
  echo "===== $HOST ====="

  if [ "$HOST" = "127.0.0.1" ] ||
     [ "$HOST" = "localhost" ]; then

    echo "hostname=$(hostname)"

    command -v mpirun

    test -x "$MPI_DIR/hello_mpi"
    test -x "$MPI_DIR/sleep_mpi"

    echo "PASS: local"

    continue
  fi

  if ssh \
    -o BatchMode=yes \
    -o ConnectTimeout=5 \
    "${SSH_USER}@${HOST}" \
    "
      echo hostname=\$(hostname)
      command -v mpirun
      test -x '$MPI_DIR/hello_mpi'
      test -x '$MPI_DIR/sleep_mpi'
    "
  then

    echo "PASS: $HOST"

  else

    echo "FAIL: $HOST"
    FAIL=1

  fi

done < "$HPC_MPI_HOSTFILE"

echo
echo "hosts=$HOSTS"

if [ "$FAIL" -ne 0 ]; then
  echo "FAIL: SSH/MPI preflight"
  exit 1
fi

echo "PASS: SSH/MPI preflight"
