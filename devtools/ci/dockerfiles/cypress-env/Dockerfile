FROM cypress/included:3.6.0
ENTRYPOINT [ "/bin/sh" ]
RUN apt-get update && apt-get install -y --no-install-recommends postgresql postgresql-contrib && rm -rf /var/lib/apt/lists/*
ENV PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@
RUN mkdir -p ${PGDATA} /run/postgresql &&\
    chown postgres ${PGDATA} /run/postgresql &&\
    su postgres -c "/usr/lib/postgresql/9.6/bin/initdb $PGDATA" &&\
    echo "host all  all    0.0.0.0/0  md5" >> $PGDATA/pg_hba.conf &&\
    echo "listen_addresses='*'" >> $PGDATA/postgresql.conf

COPY start_postgres.sh /usr/bin/start_postgres
