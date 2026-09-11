#!/usr/bin/env bash

set -u

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." &&
  pwd
)"

OUT_DIR="$ROOT_DIR/security/pki/runtime"

SERVER_LAN_IP="${SERVER_LAN_IP:-192.168.10.13}"

mkdir -p "$OUT_DIR"

chmod 700 "$OUT_DIR"

echo "[PKI] IP LAN incluida en certificados:"
echo "      $SERVER_LAN_IP"

echo
echo "[PKI] Generando CA..."

openssl genrsa \
  -out "$OUT_DIR/ca.key" \
  4096

chmod 600 \
  "$OUT_DIR/ca.key"

openssl req \
  -x509 \
  -new \
  -sha256 \
  -days 3650 \
  -key "$OUT_DIR/ca.key" \
  -out "$OUT_DIR/ca.crt" \
  -subj "/C=CO/ST=Santander/O=UPB-CIENTIFICA/OU=Security/CN=UPB-CIENTIFICA Development CA"

echo
echo "[PKI] Generando certificado Sync..."

openssl genrsa \
  -out "$OUT_DIR/sync.key" \
  3072

chmod 600 \
  "$OUT_DIR/sync.key"

cat > "$OUT_DIR/sync.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,IP:127.0.0.1,IP:${SERVER_LAN_IP}
EOF

openssl req \
  -new \
  -key "$OUT_DIR/sync.key" \
  -out "$OUT_DIR/sync.csr" \
  -subj "/C=CO/ST=Santander/O=UPB-CIENTIFICA/OU=Sync/CN=localhost"

openssl x509 \
  -req \
  -sha256 \
  -days 825 \
  -in "$OUT_DIR/sync.csr" \
  -CA "$OUT_DIR/ca.crt" \
  -CAkey "$OUT_DIR/ca.key" \
  -CAcreateserial \
  -out "$OUT_DIR/sync.crt" \
  -extfile "$OUT_DIR/sync.ext"

rm -f \
  "$OUT_DIR/sync.csr" \
  "$OUT_DIR/sync.ext"

echo
echo "[PKI] Verificando certificado..."

openssl verify \
  -CAfile "$OUT_DIR/ca.crt" \
  "$OUT_DIR/sync.crt"

echo
echo "[PKI] SAN Sync:"

openssl x509 \
  -in "$OUT_DIR/sync.crt" \
  -noout \
  -ext subjectAltName

echo
echo "[PKI] Generación terminada."
