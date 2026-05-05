#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  SELECT 'CREATE DATABASE maze_host'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'maze_host')\gexec
EOSQL
