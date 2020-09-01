FROM golang:1.14.4-alpine3.12

ENV PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@
RUN apk --no-cache add git nodejs yarn make postgresql postgresql-contrib
RUN mkdir -p ${PGDATA} /run/postgresql /var/log/postgresql &&\
    chown postgres ${PGDATA} /run/postgresql /var/log/postgresql &&\
    su postgres -c "initdb $PGDATA" &&\
    echo "host all  all    0.0.0.0/0  md5" >> $PGDATA/pg_hba.conf &&\
    echo "listen_addresses='*'" >> $PGDATA/postgresql.conf &&\
    echo "fsync = off" >> $PGDATA/postgresql.conf &&\
    echo "full_page_writes = off" >> $PGDATA/postgresql.conf

COPY start_postgres.sh /usr/bin/start_postgres
COPY stop_postgres.sh /usr/bin/stop_postgres
