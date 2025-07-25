FROM docker.io/library/golang:1.24.5-bookworm

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

COPY start_postgres.sh /opt/start_postgres.sh
COPY stop_postgres.sh /opt/stop_postgres.sh
COPY shell-runner.c /opt/shell-runner.c
RUN gcc /opt/shell-runner.c -o /opt/shell-runner
RUN chown postgres:postgres /opt/shell-runner
RUN chmod +xs /opt/shell-runner
RUN ln -s /opt/shell-runner /usr/bin/start_postgres
RUN ln -s /opt/shell-runner /usr/bin/stop_postgres

# Cypress
ENV QT_X11_NO_MITSHM=1 _X11_NO_MITSHM=1 _MITSHM=0 CYPRESS_CACHE_FOLDER=/home/user/.cache/Cypress
ENV DBUS_SESSION_BUS_ADDRESS=/dev/null

ENV CI=1


# Even though we use bun, we still need node for playwright (incl. storybook)
RUN curl -fsSL https://nodejs.org/dist/v20.19.4/node-v20.19.4-linux-x64.tar.xz | tar -xJ -C /opt && \
    ln -s /opt/node-v20.19.4-linux-x64/bin/node /usr/bin/node && \
    ln -s /opt/node-v20.19.4-linux-x64/bin/npm /usr/bin/npm && \
    ln -s /opt/node-v20.19.4-linux-x64/bin/npx /usr/bin/npx


# Create a default user and switch to it
RUN useradd -m -s /bin/bash user
USER user
ENV HOME=/home/user
WORKDIR /home/user

