FROM docker.io/goalert/build-env:test-all.1 AS build
COPY / /build/
WORKDIR /build
RUN make bin/build/goalert-linux-amd64

FROM docker.io/library/alpine
RUN apk --no-cache add ca-certificates
ENV GOALERT_LISTEN :8081
EXPOSE 8081
CMD ["/usr/bin/goalert"]

COPY --from=build /build/bin/build/goalert-linux-amd64/goalert/bin/* /usr/bin/
RUN /usr/bin/goalert self-test
