#!/usr/bin/env bash

set -euo pipefail

LDAP_BASE_DN="${LDAP_BASE_DN:-dc=upb-cientifica,dc=local}"
LDAP_AUTH_READER_DN="${LDAP_AUTH_READER_DN:-uid=auth-reader,ou=services,${LDAP_BASE_DN}}"

MDB_DN="$(
    sudo ldapsearch \
        -Q \
        -Y EXTERNAL \
        -H ldapi:/// \
        -b cn=config \
        -LLL \
        '(&(objectClass=olcDatabaseConfig)(olcDatabase=*mdb))' \
        dn |
    awk '/^dn: / {
        sub(/^dn: /, "")
        print
        exit
    }'
)"

if [[ -z "${MDB_DN}" ]]; then
    echo "ERROR: no se encontró la base MDB en cn=config." >&2
    exit 1
fi

ACL_FILE="$(mktemp)"

cleanup() {
    rm -f "${ACL_FILE}"
}

trap cleanup EXIT

cat > "${ACL_FILE}" <<LDIF
dn: ${MDB_DN}
changetype: modify
replace: olcAccess
olcAccess: {0}to * by dn.exact="gidNumber=0+uidNumber=0,cn=peercred,cn=external,cn=auth" manage by * break
olcAccess: {1}to attrs=userPassword by self write by anonymous auth by * none
olcAccess: {2}to attrs=shadowLastChange by self write by dn.exact="${LDAP_AUTH_READER_DN}" read by * none
olcAccess: {3}to * by dn.exact="${LDAP_AUTH_READER_DN}" read by * none
LDIF

sudo ldapmodify \
    -Q \
    -Y EXTERNAL \
    -H ldapi:/// \
    -f "${ACL_FILE}"

echo "ACL LDAP endurecidas correctamente."
