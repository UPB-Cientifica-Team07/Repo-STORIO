#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ANDROID_DIR="$ROOT/clients/android"

cd "$ROOT"

echo "========================================"
echo " ANDROID FILE SYNC - VERIFICACIÓN"
echo "========================================"

echo
echo "[1/8] Proto Android vs servidor"

if ! cmp -s \
    services/sync-service/proto/sync.proto \
    clients/android/app/src/main/proto/sync.proto
then
    echo "ERROR: sync.proto Android difiere del servidor."
    exit 1
fi

echo "OK: protocolos idénticos"

echo
echo "[2/8] AcknowledgeChanges"

grep -q \
    'rpc AcknowledgeChanges' \
    clients/android/app/src/main/proto/sync.proto

grep -q \
    'fun acknowledgeChanges' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/grpc/SyncGrpcClient.kt

grep -q \
    'grpcClient.acknowledgeChanges' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/sync/SyncEngine.kt

echo "OK: ACK implementado"

echo
echo "[3/8] Credenciales embebidas"

if grep -RniE \
    'tercero|123456|upb_dev_2026|const val USERNAME|const val PASSWORD' \
    clients/android/app/src/main \
    --exclude='*.bak'
then
    echo "ERROR: se encontraron credenciales embebidas."
    exit 1
fi

echo "OK: sin credenciales de prueba embebidas"

echo
echo "[4/8] CredentialStore"

grep -q \
    'class CredentialStore' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/config/CredentialStore.kt

grep -q \
    'CredentialStore' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/sync/SyncEngine.kt

echo "OK: configuración local disponible"

echo
echo "[5/8] WorkManager"

grep -q \
    'PeriodicWorkRequestBuilder' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/UPBSyncApplication.kt

grep -q \
    'OneTimeWorkRequestBuilder' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/UPBSyncApplication.kt

grep -q \
    'NetworkType.CONNECTED' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/UPBSyncApplication.kt

echo "OK: sincronización background configurada"

echo
echo "[6/8] Política Wi-Fi/LAN"

grep -q \
    'NetworkCapabilities.TRANSPORT_WIFI' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/network/LocalNetworkPolicy.kt

grep -q \
    '192.168.10.' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/network/LocalNetworkPolicy.kt

grep -q \
    'serverReachable' \
    clients/android/app/src/main/java/co/edu/upb/cientifica/sync/network/LocalNetworkPolicy.kt

echo "OK: Wi-Fi + LAN + alcance servidor"

echo
echo "[7/8] Archivos backup"

if find clients/android \
    -type f \
    \( \
      -name '*.bak' \
      -o -name '*.pre-grpc' \
    \) \
    | grep -q .
then
    echo "ERROR: quedan backups obsoletos."
    exit 1
fi

echo "OK: sin backups obsoletos"

echo
echo "[8/8] Gradle"

cd "$ANDROID_DIR"

./gradlew \
    testDebugUnitTest \
    assembleDebug

echo
echo "========================================"
echo " ANDROID FILE SYNC: VERIFICACIÓN OK"
echo "========================================"

ls -lh \
    app/build/outputs/apk/debug/app-debug.apk
