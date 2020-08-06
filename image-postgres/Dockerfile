FROM postgres:12-alpine

COPY initdb/* /docker-entrypoint-initdb.d/
RUN chmod 777 /docker-entrypoint-initdb.d/20-user-database.sql

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["postgres"]