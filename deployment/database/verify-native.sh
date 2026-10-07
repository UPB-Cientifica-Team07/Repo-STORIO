#!/usr/bin/env bash

set -u

HOST="${POSTGRES_HOST:-127.0.0.1}"
PORT="${POSTGRES_PORT:-5434}"
DB="${POSTGRES_DB:-upb_cientifica}"
USER_NAME="${POSTGRES_USER:-upb_app}"

RC=0

echo "===== UPB-CIENTIFICA NATIVE POSTGRESQL CHECK ====="

echo "HOST=$HOST"
echo "PORT=$PORT"
echo "DATABASE=$DB"
echo "USER=$USER_NAME"

if command -v psql >/dev/null 2>&1; then
  echo "PSQL=FOUND"
else
  echo "PSQL=NOT_FOUND"
  RC=1
fi

if command -v pg_isready >/dev/null 2>&1; then
  echo "PG_ISREADY=FOUND"
else
  echo "PG_ISREADY=NOT_FOUND"
  RC=1
fi

if test -n "${POSTGRES_PASSWORD:-}"; then
  echo "POSTGRES_PASSWORD=SET"
else
  echo "POSTGRES_PASSWORD=MISSING"
  RC=1
fi

if test "$RC" -eq 0; then
  if pg_isready \
    -h "$HOST" \
    -p "$PORT" \
    -d "$DB" \
    -U "$USER_NAME" \
    >/dev/null 2>&1; then

    echo "POSTGRES_READY=PASS"

  else

    echo "POSTGRES_READY=FAIL"
    RC=1

  fi
fi

if test "$RC" -eq 0; then

  RESULT="$(
    PGPASSWORD="$POSTGRES_PASSWORD" \
    psql \
      -h "$HOST" \
      -p "$PORT" \
      -U "$USER_NAME" \
      -d "$DB" \
      -v ON_ERROR_STOP=1 \
      -At \
      -c "
SELECT
    'SERVER_PORT=' || inet_server_port();

SELECT
    'DATABASE=' || current_database();

SELECT
    'CURRENT_USER=' || current_user;

SELECT
    'REQUIRED_TABLES=' || COUNT(*)
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN (
    'usuario',
    'home',
    'archivo',
    'imagen',
    'album',
    'tag',
    'permiso_recurso',
    'sync_file',
    'sync_change',
    'sync_device_cursor',
    'nodo_hpc',
    'trabajo_hpc',
    'grid_resource'
);
"
  )"

  QUERY_RC=$?

  printf '%s\n' "$RESULT"

  if test "$QUERY_RC" -eq 0; then
    echo "DATABASE_QUERY=PASS"
  else
    echo "DATABASE_QUERY=FAIL"
    RC=1
  fi
fi

if test "$RC" -eq 0; then
  echo "NATIVE_POSTGRESQL=PASS"
else
  echo "NATIVE_POSTGRESQL=FAIL"
fi

exit "$RC"
