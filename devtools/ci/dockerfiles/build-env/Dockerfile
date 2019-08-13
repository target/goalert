FROM postgres:11

RUN apt-get update && apt-get install -y build-essential nginx curl xvfb libgtk2.0-0 libnotify-dev libgconf-2-4 libnss3 libxss1 libasound2 libpng-dev git && rm -rf /var/lib/apt/lists/*

ENV GO_VERSION=1.12.7
RUN curl -sL https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar xz -C /usr/local

ENV NODE_VERSION=10.16.2
RUN curl -sL https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz | tar xJ --strip-components 1 -C / && npm install -g yarn

ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/node_modules/.bin:/usr/lib/postgresql/11/bin:/usr/local/go/bin:/root/go/bin
