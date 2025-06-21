#!/bin/sh
set -e

echo "init-replication.sh running"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE ROLE ${REPLICATION_USER} REPLICATION LOGIN ENCRYPTED PASSWORD '${REPLICATION_PASSWORD}' VALID UNTIL 'infinity';
  SELECT * FROM pg_create_physical_replication_slot('replica_slot');
EOSQL

cat >> "$PGDATA"/postgresql.conf <<-EOF

wal_level = replica
max_wal_senders = 10
wal_keep_size = '64MB'
listen_addresses = '*'
EOF

cat >> "$PGDATA"/pg_hba.conf <<-EOF
host replication ${REPLICATION_USER} 0.0.0.0/0 md5
EOF