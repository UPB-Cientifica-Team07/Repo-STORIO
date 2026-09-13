#!/usr/bin/env bash

set -euo pipefail

OUTPUT="${1:-services/hpc-service/lab/hosts.generated}"

if [ "$#" -lt 2 ]; then
    echo "Uso: $0 <output> <host[:slots]> [host[:slots] ...]"
    echo
    echo "Ejemplo:"
    echo "  $0 /tmp/upb-hpc-hosts 192.168.10.101:2 192.168.10.102:2"
    exit 1
fi

shift

: > "$OUTPUT"

TOTAL_SLOTS=0

for entry in "$@"; do

    HOST="${entry%%:*}"

    if [ "$HOST" = "$entry" ]; then
        SLOTS=1
    else
        SLOTS="${entry##*:}"
    fi

    if ! [[ "$SLOTS" =~ ^[1-9][0-9]*$ ]]; then
        echo "ERROR: slots invalidos para $entry"
        exit 1
    fi

    printf '%s slots=%s\n' \
        "$HOST" \
        "$SLOTS" \
        >> "$OUTPUT"

    TOTAL_SLOTS=$((TOTAL_SLOTS + SLOTS))
done

echo "===== MPI HOSTFILE ====="
cat "$OUTPUT"
echo
echo "hosts=$(wc -l < "$OUTPUT")"
echo "slots=$TOTAL_SLOTS"
echo "output=$OUTPUT"
