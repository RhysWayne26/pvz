#!/bin/sh
set -e
echo "setup-slave.sh: waiting for master"

until pg_isready -h postgres-master -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"; do
  echo "  master not ready…"
  sleep 2
done

echo "master is up — doing basebackup"
rm -rf "${PGDATA:?}"/*

export PGPASSWORD="${REPLICATION_PASSWORD}"
pg_basebackup \
  -h postgres-master \
  -D "${PGDATA}" \
  -U "${REPLICATION_USER}" \
  -v -P \
  --wal-method=stream

echo "configuring standby"
touch "${PGDATA}/standby.signal"
cat >> "${PGDATA}/postgresql.auto.conf" <<-EOF
primary_conninfo = 'host=postgres-master port=5432 user=${REPLICATION_USER} password=${REPLICATION_PASSWORD}'
primary_slot_name = 'replica_slot'
hot_standby = on
EOF

echo "launching postgres"
exec docker-entrypoint.sh postgres
