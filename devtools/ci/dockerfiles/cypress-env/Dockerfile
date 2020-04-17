FROM cypress/included:4.4.0
ENTRYPOINT [ "/bin/sh" ]
RUN apt-get update && apt-get install -y --no-install-recommends postgresql postgresql-contrib && rm -rf /var/lib/apt/lists/*
ENV PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@
RUN mkdir -p ${PGDATA} /run/postgresql /var/log/postgresql &&\
    chown postgres ${PGDATA} /run/postgresql /var/log/postgresql &&\
    su postgres -c "/usr/lib/postgresql/11/bin/initdb $PGDATA" &&\
    echo "host all  all    0.0.0.0/0  md5" >> $PGDATA/pg_hba.conf &&\
    echo "listen_addresses='*'" >> $PGDATA/postgresql.conf &&\
    echo "fsync = off" >> $PGDATA/postgresql.conf &&\
    echo "full_page_writes = off" >> $PGDATA/postgresql.conf

COPY start_postgres.sh /usr/bin/start_postgres
COPY stop_postgres.sh /usr/bin/stop_postgres
