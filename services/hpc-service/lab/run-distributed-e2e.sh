#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

: "${TOKEN_HPC:?Falta TOKEN_HPC}"

COORDINATOR_IP="${COORDINATOR_IP:-127.0.0.1}"
COORDINATOR_PORT="${COORDINATOR_PORT:-1100}"
HPC_PROCESSES="${HPC_PROCESSES:-2}"
HPC_PROGRAM="${HPC_PROGRAM:-hello_mpi}"
EXPECT_HOSTS="${HPC_EXPECT_MIN_HOSTS:-1}"
JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"
EVIDENCE_DIR="${HPC_EVIDENCE_DIR:-/tmp/upb-hpc-evidence}"

mkdir -p "$EVIDENCE_DIR"

STAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT="$EVIDENCE_DIR/job-$STAMP.log"

echo "===== HPC E2E ====="
echo "coordinator=${COORDINATOR_IP}:${COORDINATOR_PORT}"
echo "program=$HPC_PROGRAM"
echo "processes=$HPC_PROCESSES"
echo "expected_hosts=$EXPECT_HOSTS"

echo
echo "===== HEALTH BEFORE ====="

"$ROOT_DIR/services/hpc-service/lab/health-check.sh"

echo
echo "===== SUBMIT JOB ====="

java \
  -cp "$ROOT_DIR/services/hpc-service/bin:$JDBC_JAR" \
  hpc.client.JobClient \
  "$COORDINATOR_IP" \
  "$COORDINATOR_PORT" \
  "$TOKEN_HPC" \
  "$HPC_PROCESSES" \
  "$HPC_PROGRAM" \
  | tee "$OUTPUT"

echo
echo "===== ASSERT SUCCESS ====="

if grep -q 'Success: true' "$OUTPUT"; then
  echo "PASS: Job success"
else
  echo "FAIL: Job no exitoso"
  exit 1
fi

DISTINCT_HOSTS="$(
  sed -n \
    -e 's/.*ejecutándose en \(.*\)$/\1/p' \
    -e 's/.*iniciado en \(.*\)$/\1/p' \
    "$OUTPUT" \
  | sort -u \
  | tee "$EVIDENCE_DIR/hosts-$STAMP.txt" \
  | wc -l
)"

echo "distinct_hosts=$DISTINCT_HOSTS"

if [ "$DISTINCT_HOSTS" -lt "$EXPECT_HOSTS" ]; then
  echo "FAIL: hosts MPI observados insuficientes"
  exit 1
fi

echo "PASS: host count MPI"

echo
echo "===== HEALTH AFTER ====="

"$ROOT_DIR/services/hpc-service/lab/health-check.sh"

echo
echo "PASS: HPC E2E completo"
