#!/usr/bin/env bash

set -u

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." &&
  pwd
)"

OUT_DIR="$ROOT_DIR/security/pki/runtime"

SERVER_LAN_IP="${SERVER_LAN_IP:-192.168.10.13}"

CA_CERT="$OUT_DIR/ca.crt"
CA_KEY="$OUT_DIR/ca.key"

AUTH_KEY="$OUT_DIR/auth.key"
AUTH_CSR="$OUT_DIR/auth.csr"
AUTH_CERT="$OUT_DIR/auth.crt"
AUTH_EXT="$OUT_DIR/auth.ext"
AUTH_P12="$OUT_DIR/auth.p12"

if [ ! -f "$CA_CERT" ]; then
  echo "ERROR: no existe $CA_CERT"
  exit 1
fi

if [ ! -f "$CA_KEY" ]; then
  echo "ERROR: no existe $CA_KEY"
  exit 1
fi

if [ -z "${AUTH_TLS_KEYSTORE_PASSWORD:-}" ]; then
  echo "ERROR: AUTH_TLS_KEYSTORE_PASSWORD no está definido"
  exit 1
fi

echo "[PKI] Reutilizando CA existente"
echo "[PKI] IP LAN Auth: $SERVER_LAN_IP"

openssl genrsa \
  -out "$AUTH_KEY" \
  3072

chmod 600 \
  "$AUTH_KEY"

cat > "$AUTH_EXT" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,IP:127.0.0.1,IP:${SERVER_LAN_IP}
EOF

openssl req \
  -new \
  -key "$AUTH_KEY" \
  -out "$AUTH_CSR" \
  -subj "/C=CO/ST=Santander/O=UPB-CIENTIFICA/OU=Auth/CN=localhost"

openssl x509 \
  -req \
  -sha256 \
  -days 825 \
  -in "$AUTH_CSR" \
  -CA "$CA_CERT" \
  -CAkey "$CA_KEY" \
  -CAcreateserial \
  -out "$AUTH_CERT" \
  -extfile "$AUTH_EXT"

openssl pkcs12 \
  -export \
  -name upb-auth \
  -inkey "$AUTH_KEY" \
  -in "$AUTH_CERT" \
  -certfile "$CA_CERT" \
  -out "$AUTH_P12" \
  -passout env:AUTH_TLS_KEYSTORE_PASSWORD

chmod 600 \
  "$AUTH_P12"

rm -f \
  "$AUTH_CSR" \
  "$AUTH_EXT"

echo
echo "[PKI] Verificando certificado Auth..."

openssl verify \
  -CAfile "$CA_CERT" \
  "$AUTH_CERT"

echo
echo "[PKI] SAN Auth..."

openssl x509 \
  -in "$AUTH_CERT" \
  -noout \
  -ext subjectAltName

echo
echo "[PKI] Auth PKCS12 generado"
