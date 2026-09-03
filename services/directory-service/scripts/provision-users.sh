#!/usr/bin/env bash

set -euo pipefail

LDAP_URL="${LDAP_URL:-ldap://127.0.0.1:389}"
LDAP_BASE_DN="${LDAP_BASE_DN:-dc=upb-cientifica,dc=local}"
LDAP_ADMIN_DN="${LDAP_ADMIN_DN:-cn=admin,${LDAP_BASE_DN}}"

if [[ -z "${LDAP_ADMIN_PASSWORD:-}" ]]; then
    echo "Falta LDAP_ADMIN_PASSWORD" >&2
    exit 1
fi

if [[ -z "${LDAP_DEFAULT_USER_PASSWORD:-}" ]]; then
    echo "Falta LDAP_DEFAULT_USER_PASSWORD" >&2
    exit 1
fi

ROOT_DIR="$(
    cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &&
    pwd
)"

DIRECTORY_DIR="${ROOT_DIR}/services/directory-service"
LOCAL_LDIF="${DIRECTORY_DIR}/ldif/03-users.local.ldif"

SAMUEL_HASH="$(
    slappasswd -s "${LDAP_DEFAULT_USER_PASSWORD}"
)"

PRUEBA_HASH="$(
    slappasswd -s "${LDAP_DEFAULT_USER_PASSWORD}"
)"

TERCERO_HASH="$(
    slappasswd -s "${LDAP_DEFAULT_USER_PASSWORD}"
)"

cat > "${LOCAL_LDIF}" <<EOF
dn: uid=samuel,ou=people,${LDAP_BASE_DN}
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
uid: samuel
cn: Samuel
sn: Samuel
displayName: Samuel
employeeNumber: user-001
description: Usuario ADMIN de UPB-CIENTIFICA
userPassword: ${SAMUEL_HASH}

dn: uid=prueba,ou=people,${LDAP_BASE_DN}
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
uid: prueba
cn: Prueba
sn: Prueba
displayName: Prueba
employeeNumber: user-002
description: Usuario estándar de UPB-CIENTIFICA
userPassword: ${PRUEBA_HASH}

dn: uid=tercero,ou=people,${LDAP_BASE_DN}
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
uid: tercero
cn: Tercero
sn: Tercero
displayName: Tercero
employeeNumber: user-003
description: Usuario estándar de UPB-CIENTIFICA
userPassword: ${TERCERO_HASH}
EOF

ldapadd \
    -x \
    -H "${LDAP_URL}" \
    -D "${LDAP_ADMIN_DN}" \
    -w "${LDAP_ADMIN_PASSWORD}" \
    -f "${LOCAL_LDIF}"

echo
echo "Asignando roles mediante grupos LDAP..."

ldapmodify \
    -x \
    -H "${LDAP_URL}" \
    -D "${LDAP_ADMIN_DN}" \
    -w "${LDAP_ADMIN_PASSWORD}" \
    -f "${DIRECTORY_DIR}/ldif/04-group-members.ldif"

echo
echo "Usuarios LDAP aprovisionados y roles asignados."
