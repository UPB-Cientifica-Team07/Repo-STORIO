#!/usr/bin/env bash

set -euo pipefail

echo "===== UPB-CIENTIFICA HPC LAB PREFLIGHT ====="

FAIL=0

check_command() {

    local command_name="$1"

    if command -v "$command_name" >/dev/null 2>&1; then
        echo "PASS: $command_name -> $(command -v "$command_name")"
    else
        echo "FAIL: falta $command_name"
        FAIL=1
    fi
}

check_command java
check_command javac
check_command mpirun
check_command mpicc
check_command ssh
check_command scp
check_command hostname
check_command ip
check_command curl
check_command nc

echo
echo "===== JAVA ====="
java -version 2>&1 | head -n 3 || true

echo
echo "===== OPENMPI ====="
mpirun --version 2>/dev/null | head -n 5 || true

echo
echo "===== HOST ====="
echo "hostname=$(hostname)"

ip -brief address \
    | awk '$1 != "lo" {print}'

echo
echo "===== JDBC ====="

JDBC_JAR="${JDBC_JAR:-/usr/share/java/postgresql.jar}"

if [ -r "$JDBC_JAR" ]; then
    echo "PASS: JDBC_JAR=$JDBC_JAR"
else
    echo "FAIL: PostgreSQL JDBC no encontrado en $JDBC_JAR"
    FAIL=1
fi

echo
echo "===== MPI BINARIES ====="

for binary in \
    services/hpc-service/mpi/hello_mpi \
    services/hpc-service/mpi/sleep_mpi
do
    if [ -x "$binary" ]; then
        echo "PASS: $binary"
    else
        echo "FAIL: $binary no existe o no es ejecutable"
        FAIL=1
    fi
done

echo
echo "===== RESULT ====="

if [ "$FAIL" -eq 0 ]; then
    echo "PASS: nodo preparado para HPC"
    exit 0
fi

echo "FAIL: nodo no esta listo"
exit 1
