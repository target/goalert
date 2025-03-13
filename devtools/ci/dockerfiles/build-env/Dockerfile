FROM docker.io/library/golang:1.24.0-bullseye

RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common \
    build-essential \
    postgresql-13 postgresql-client-13 \
    dbus-x11 xvfb && \
    rm -rf /var/lib/apt/lists/*

ADD https://nodejs.org/dist/v20.11.0/node-v20.11.0-linux-x64.tar.xz /tmp/node.tar.xz
RUN tar -xvf /tmp/node.tar.xz -C /usr/local --strip-components=1 && rm /tmp/node.tar.xz

# yarn
RUN npm install -g yarn

# google chrome
RUN \
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google.list && \
    apt-get update && \
    apt-get install -y google-chrome-stable && \
    rm -rf /var/lib/apt/lists/*

# Postgres
ENV PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@?client_encoding=UTF8
RUN mkdir -p ${PGDATA} /run/postgresql /var/log/postgresql &&\
    chown postgres ${PGDATA} /run/postgresql /var/log/postgresql &&\
    su postgres -c "/usr/lib/postgresql/13/bin/initdb $PGDATA" &&\
    echo "host all  all    0.0.0.0/0  md5" >> $PGDATA/pg_hba.conf &&\
    echo "listen_addresses='*'" >> $PGDATA/postgresql.conf &&\
    echo "fsync = off" >> $PGDATA/postgresql.conf &&\
    echo "full_page_writes = off" >> $PGDATA/postgresql.conf
COPY start_postgres.sh /usr/bin/start_postgres
COPY stop_postgres.sh /usr/bin/stop_postgres

# Cypress
ENV QT_X11_NO_MITSHM=1 _X11_NO_MITSHM=1 _MITSHM=0 CYPRESS_CACHE_FOLDER=/root/.cache/Cypress
ENV DBUS_SESSION_BUS_ADDRESS=/dev/null

ENV CI=1
