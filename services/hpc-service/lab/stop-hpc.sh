#!/usr/bin/env bash

set -euo pipefail

LOG_DIR="${HPC_LAB_LOG_DIR:-/tmp/upb-hpc-lab}"

if [ ! -d "$LOG_DIR" ]; then
    echo "INFO: no existe $LOG_DIR"
    exit 0
fi

echo "===== STOP HPC LAB PROCESSES ====="

for PID_FILE in "$LOG_DIR"/*.pid; do

    [ -f "$PID_FILE" ] || continue

    PID="$(
        cat "$PID_FILE" 2>/dev/null || true
    )"

    NAME="$(
        basename "$PID_FILE" .pid
    )"

    if [ -n "$PID" ] &&
       kill -0 "$PID" 2>/dev/null; then

        echo "Stopping $NAME PID=$PID"

        kill "$PID"

        for _ in $(seq 1 30); do

            if ! kill -0 "$PID" 2>/dev/null; then
                break
            fi

            sleep 0.1
        done

        if kill -0 "$PID" 2>/dev/null; then
            echo "WARN: $NAME no termino con SIGTERM"
        else
            echo "PASS: $NAME detenido"
        fi

    else
        echo "INFO: $NAME ya no esta activo"
    fi

    rm -f "$PID_FILE"
done
