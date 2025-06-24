FROM postgres:15-alpine
COPY init-replication.sh /docker-entrypoint-initdb.d/init-replication.sh
RUN chmod +x /docker-entrypoint-initdb.d/init-replication.sh
