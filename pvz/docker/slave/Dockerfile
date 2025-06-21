FROM postgres:15-alpine

COPY setup-slave.sh /usr/local/bin/setup-slave.sh
RUN chmod +x /usr/local/bin/setup-slave.sh

ENTRYPOINT ["/usr/local/bin/setup-slave.sh"]