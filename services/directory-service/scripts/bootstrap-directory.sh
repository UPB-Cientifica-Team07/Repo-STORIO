#!/usr/bin/env bash

set -euo pipefail

LDAP_URL="${LDAP_URL:-ldaps://127.0.0.1:636}"
LDAP_BASE_DN="${LDAP_BASE_DN:-dc=upb-cientifica,dc=local}"
LDAP_ADMIN_DN="${LDAP_ADMIN_DN:-cn=admin,${LDAP_BASE_DN}}"

if [[ -z "${LDAP_ADMIN_PASSWORD:-}" ]]; then
    echo "ERROR: LDAP_ADMIN_PASSWORD es obligatoria." >&2
    exit 1
fi

ROOT_DIR="$(
    cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
    pwd
)"

LDAP_TLS_CA_FILE="${LDAP_TLS_CA_FILE:-${ROOT_DIR}/security/pki/upb_dev_ca.crt}"

if [[ ! -r "${LDAP_TLS_CA_FILE}" ]]; then
    echo "ERROR: CA LDAP no legible: ${LDAP_TLS_CA_FILE}" >&2
    exit 1
fi

export LDAPTLS_CACERT="${LDAP_TLS_CA_FILE}"

DIRECTORY_DIR="${ROOT_DIR}/services/directory-service"

echo "==================================="
echo " DIRECTORY SERVICE BOOTSTRAP"
echo "==================================="
echo "LDAP URL: ${LDAP_URL}"
echo "Base DN: ${LDAP_BASE_DN}"
echo

echo "[1/2] Creando estructura..."

ldapadd \
    -x \
    -H "${LDAP_URL}" \
    -D "${LDAP_ADMIN_DN}" \
    -w "${LDAP_ADMIN_PASSWORD}" \
    -f "${DIRECTORY_DIR}/ldif/01-structure.ldif"

echo
echo "[2/2] Creando grupos..."

ldapadd \
    -x \
    -H "${LDAP_URL}" \
    -D "${LDAP_ADMIN_DN}" \
    -w "${LDAP_ADMIN_PASSWORD}" \
    -f "${DIRECTORY_DIR}/ldif/02-groups.ldif"

echo
echo "==================================="
echo " ESTRUCTURA LDAP CREADA"
echo "==================================="
echo
echo "Siguiente paso:"
echo "  scripts/provision-users.sh"
