#!/usr/bin/env bash

set -euo pipefail

: "${HPC_MPI_HOSTFILE:?Falta HPC_MPI_HOSTFILE}"

MPI_DIR="${HPC_REMOTE_MPI_DIR:-/tmp/upb-hpc-mpi}"

TOTAL_SLOTS="$(
  awk '
    {
      for (i = 1; i <= NF; i++) {
        if ($i ~ /^slots=/) {
          split($i, a, "=")
          total += a[2]
        }
      }
    }
    END {
      print total + 0
    }
  ' "$HPC_MPI_HOSTFILE"
)"

MPI_PROCESSES="${HPC_PROCESSES:-$TOTAL_SLOTS}"

if [ "$MPI_PROCESSES" -lt 1 ]; then
  echo "ERROR: procesos MPI invalidos"
  exit 1
fi

echo "===== DIRECT MPI TEST ====="
echo "processes=$MPI_PROCESSES"

mpirun \
  --hostfile "$HPC_MPI_HOSTFILE" \
  --map-by slot \
  -np "$MPI_PROCESSES" \
  "$MPI_DIR/hello_mpi" \
  | tee /tmp/upb-hpc-direct-mpi.log

DISTINCT="$(
  sed -n \
    's/.*ejecutándose en \(.*\)$/\1/p' \
    /tmp/upb-hpc-direct-mpi.log \
  | sort -u \
  | tee /tmp/upb-hpc-direct-hosts.txt \
  | wc -l
)"

echo "distinct_hosts=$DISTINCT"
echo "PASS: direct MPI completed"
