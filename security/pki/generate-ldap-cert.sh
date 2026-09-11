#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." &&
  pwd
)"

PKI_DIR="$ROOT_DIR/security/pki"
RUNTIME_DIR="$PKI_DIR/runtime"

CA_CERT="$RUNTIME_DIR/ca.crt"
CA_KEY="$RUNTIME_DIR/ca.key"

LDAP_KEY="$RUNTIME_DIR/ldap.key"
LDAP_CSR="$RUNTIME_DIR/ldap.csr"
LDAP_CERT="$RUNTIME_DIR/ldap.crt"
LDAP_EXT="$RUNTIME_DIR/ldap.ext"

if [[ ! -f "$CA_CERT" || ! -f "$CA_KEY" ]]; then
  echo "ERROR: CA existente no encontrada."
  echo "No se generará una CA nueva."
  exit 1
fi

mkdir -p "$RUNTIME_DIR"

umask 077

openssl genrsa \
  -out "$LDAP_KEY" \
  3072

openssl req \
  -new \
  -key "$LDAP_KEY" \
  -out "$LDAP_CSR" \
  -subj "/C=CO/ST=Santander/O=UPB-CIENTIFICA/OU=Directory/CN=localhost"

cat > "$LDAP_EXT" <<'EXT'
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,IP:127.0.0.1,IP:192.168.10.13,IP:192.168.10.18
EXT

openssl x509 \
  -req \
  -in "$LDAP_CSR" \
  -CA "$CA_CERT" \
  -CAkey "$CA_KEY" \
  -CAcreateserial \
  -out "$LDAP_CERT" \
  -days 825 \
  -sha256 \
  -extfile "$LDAP_EXT"

chmod 600 \
  "$LDAP_KEY"

chmod 644 \
  "$LDAP_CERT"

rm -f \
  "$LDAP_CSR" \
  "$LDAP_EXT"

echo "LDAP certificate generated:"
echo "$LDAP_CERT"
