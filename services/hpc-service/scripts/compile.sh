#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
  pwd
)"

HPC_DIR="$ROOT_DIR/services/hpc-service"
BIN_DIR="$HPC_DIR/bin"
MPI_DIR="$HPC_DIR/mpi"

JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"

if [[ ! -f "$JDBC_JAR" ]]; then
  echo "ERROR: PostgreSQL JDBC no encontrado:"
  echo "  $JDBC_JAR"
  exit 1
fi

command -v javac >/dev/null 2>&1 || {
  echo "ERROR: javac no está instalado"
  exit 1
}

command -v mpicc >/dev/null 2>&1 || {
  echo "ERROR: mpicc no está instalado"
  exit 1
}

rm -rf "$BIN_DIR"
mkdir -p "$BIN_DIR"

echo "[HPC] Compilando Java..."

JAVA_SOURCES=()

while IFS= read -r -d '' source; do
  JAVA_SOURCES+=("$source")
done < <(
  find "$HPC_DIR/src" \
    -type f \
    -name '*.java' \
    -print0
)

if [[ ${#JAVA_SOURCES[@]} -eq 0 ]]; then
  echo "ERROR: no se encontraron fuentes Java"
  exit 1
fi

javac \
  -cp "$JDBC_JAR" \
  -d "$BIN_DIR" \
  "${JAVA_SOURCES[@]}"

echo "[HPC] Compilando programas MPI..."

mpicc \
  "$MPI_DIR/hello_mpi.c" \
  -o "$MPI_DIR/hello_mpi"

mpicc \
  "$MPI_DIR/sleep_mpi.c" \
  -o "$MPI_DIR/sleep_mpi"

echo
echo "HPC Service compilado correctamente."
