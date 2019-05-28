FROM postgres:11

RUN apt-get update && apt-get install -y build-essential nginx curl xvfb libgtk2.0-0 libnotify-dev libgconf-2-4 libnss3 libxss1 libasound2 libpng-dev git && rm -rf /var/lib/apt/lists/*
RUN curl -s https://nodejs.org/dist/v10.15.3/node-v10.15.3-linux-x64.tar.xz | tar xJ --strip-components 1 -C /
RUN curl -s https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz | tar xz -C /usr/local
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/node_modules/.bin:/usr/lib/postgresql/9.6/bin:/usr/local/go/bin:/root/go/bin
RUN npm install -g yarn
