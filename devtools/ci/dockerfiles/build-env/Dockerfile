FROM docker.io/library/golang:1.24.2-bookworm

RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    zip \
    unzip \
    gnupg-agent \
    gnupg2 lsb-release \
    software-properties-common \
    build-essential \
    dbus-x11 xvfb && \
    rm -rf /var/lib/apt/lists/*

# Postgres
RUN mkdir -p /usr/share/postgresql-common/pgdg && curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail https://www.postgresql.org/media/keys/ACCC4CF8.asc

RUN echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt bookworm-pgdg main" > /etc/apt/sources.list.d/postgresql.list

RUN apt-get update && apt-get install -y \
    postgresql-13 postgresql-client-13 \
    postgresql-14 postgresql-client-14 \
    postgresql-15 postgresql-client-15 \
    postgresql-16 postgresql-client-16 \
    postgresql-17 postgresql-client-17 \
    && rm -rf /var/lib/apt/lists/*

# google chrome
RUN \
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google.list && \
    apt-get update && \
    apt-get install -y google-chrome-stable && \
    rm -rf /var/lib/apt/lists/*

# Postgres
ENV PGUSER=postgres DB_URL=postgresql://postgres@?client_encoding=UTF8&sslmode=disable
COPY setup_postgres.sh /usr/bin/setup_postgres
RUN setup_postgres 13 14 15 16 17

COPY start_postgres.sh /usr/bin/start_postgres
COPY stop_postgres.sh /usr/bin/stop_postgres

# Cypress
ENV QT_X11_NO_MITSHM=1 _X11_NO_MITSHM=1 _MITSHM=0 CYPRESS_CACHE_FOLDER=/root/.cache/Cypress
ENV DBUS_SESSION_BUS_ADDRESS=/dev/null

ENV CI=1
