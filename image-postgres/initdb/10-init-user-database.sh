#!/bin/sh

echo "-------
Generating 20-user-database.sql"

echo "
CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASS';
CREATE DATABASE $DB_NAME;
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" >> /docker-entrypoint-initdb.d/20-user-database.sql

echo "Result file:"
cat /docker-entrypoint-initdb.d/20-user-database.sql

echo "-------"
