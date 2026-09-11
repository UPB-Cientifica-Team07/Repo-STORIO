#!/usr/bin/env bash

set -euo pipefail

LDAP_BASE_DN="${LDAP_BASE_DN:-dc=upb-cientifica,dc=local}"
LDAP_AUTH_READER_DN="${LDAP_AUTH_READER_DN:-uid=auth-reader,ou=services,${LDAP_BASE_DN}}"

if [[ -z "${LDAP_AUTH_READER_PASSWORD:-}" ]]; then
    echo "ERROR: LDAP_AUTH_READER_PASSWORD es obligatoria." >&2
    exit 1
fi

PW_FILE="$(mktemp)"

cleanup() {
    rm -f "${PW_FILE}"
}

trap cleanup EXIT

chmod 600 "${PW_FILE}"

printf '%s' \
    "${LDAP_AUTH_READER_PASSWORD}" \
    > "${PW_FILE}"

READER_HASH="$(
    slappasswd \
        -T "${PW_FILE}"
)"

if sudo ldapsearch \
    -Q \
    -Y EXTERNAL \
    -H ldapi:/// \
    -b "${LDAP_AUTH_READER_DN}" \
    -s base \
    -LLL \
    dn \
    >/dev/null \
    2>&1
then

    sudo ldapmodify \
        -Q \
        -Y EXTERNAL \
        -H ldapi:/// <<LDIF
dn: ${LDAP_AUTH_READER_DN}
changetype: modify
replace: userPassword
userPassword: ${READER_HASH}
LDIF

    echo "auth-reader existente: contraseña actualizada."

else

    sudo ldapadd \
        -Q \
        -Y EXTERNAL \
        -H ldapi:/// <<LDIF
dn: ${LDAP_AUTH_READER_DN}
objectClass: top
objectClass: account
objectClass: simpleSecurityObject
uid: auth-reader
description: Cuenta técnica de solo lectura para Auth Service
userPassword: ${READER_HASH}
LDIF

    echo "auth-reader creado correctamente."
fi

unset READER_HASH
