#!/bin/bash
set -e

# Wait for master to be ready
until PGPASSWORD=$POSTGRES_PASSWORD psql -h postgres-master -U $POSTGRES_USER -d $POSTGRES_DB -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

# Clean up data directory
rm -rf /var/lib/postgresql/data/*

# Clone master data
PGPASSWORD=replicator_password pg_basebackup -h postgres-master -D /var/lib/postgresql/data -U replicator -v -P -X stream

# Create standby signal
touch /var/lib/postgresql/data/standby.signal

# Configure connection info
echo "primary_conninfo = 'host=postgres-master port=5432 user=replicator password=replicator_password application_name=postgres-slave'" >> /var/lib/postgresql/data/postgresql.auto.conf

# Start server
exec docker-entrypoint.sh postgres
