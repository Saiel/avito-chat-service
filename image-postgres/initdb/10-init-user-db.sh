#!/bin/sh

echo \
"CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASS';
CREATE DATABASE $DB_NAME;
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" >> /docker-entrypoint-initdb.d/20-query.sql

# psql -v ON_ERROR_STOP=1 -h localhost -U "postgres" -d "$POSTGRES_DB" -f query.sql
