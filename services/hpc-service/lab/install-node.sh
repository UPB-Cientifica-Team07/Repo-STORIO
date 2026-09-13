#!/usr/bin/env bash

set -euo pipefail

MODE="${1:-check}"

echo "===== HPC NODE DEPENDENCIES ====="
echo "mode=$MODE"

PACKAGES=(
  openjdk-21-jdk
  openmpi-bin
  libopenmpi-dev
  openssh-client
  openssh-server
  libpostgresql-jdbc-java
  netcat-openbsd
)

if [ "$MODE" = "--apply" ]; then

  if ! command -v apt-get >/dev/null 2>&1; then
    echo "ERROR: apt-get no disponible"
    exit 1
  fi

  sudo apt-get update

  sudo apt-get install -y \
    "${PACKAGES[@]}"

  if command -v systemctl >/dev/null 2>&1; then
    sudo systemctl enable --now ssh 2>/dev/null || true
  fi

fi

FAIL=0

for CMD in \
  java \
  javac \
  mpirun \
  mpicc \
  ssh \
  scp \
  nc
do

  if command -v "$CMD" >/dev/null 2>&1; then
    echo "PASS: $CMD"
  else
    echo "FAIL: $CMD"
    FAIL=1
  fi

done

JDBC_JAR="${POSTGRES_JDBC_JAR:-/usr/share/java/postgresql.jar}"

if [ -r "$JDBC_JAR" ]; then
  echo "PASS: JDBC=$JDBC_JAR"
else
  echo "FAIL: JDBC no disponible: $JDBC_JAR"
  FAIL=1
fi

if [ "$FAIL" -ne 0 ]; then
  exit 1
fi

echo "PASS: nodo preparado"
