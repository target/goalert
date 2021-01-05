FROM goalert/build-env AS build

COPY Makefile /goalert/Makefile
COPY yarn.lock package.json /goalert/
COPY web/src/package.json /goalert/web/src/
WORKDIR /goalert
RUN make node_modules

COPY go.mod go.sum /goalert/
RUN go mod download

COPY . /goalert
RUN make bin/resetdb bin/goalert BUNDLE=1 BUILD_FLAGS=-trimpath

FROM alpine:3.12
RUN apk --no-cache add postgresql postgresql-contrib musl-locales ca-certificates
COPY --from=build /goalert/bin/goalert /goalert/bin/resetdb /bin/
COPY devtools/ci/dockerfiles/all-in-one/start.sh /bin/start.sh
ENV GOALERT_LISTEN :8081
EXPOSE 8081
CMD ["/bin/start.sh"]
