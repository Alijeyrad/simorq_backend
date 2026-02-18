#!/bin/bash
# Creates extra databases required by the application.
# Runs automatically inside the postgres container on first startup.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    SELECT 'CREATE DATABASE simorq_context_db'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'simorq_context_db')\gexec
    SELECT 'CREATE DATABASE simorq_casbin_db'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'simorq_casbin_db')\gexec
EOSQL
