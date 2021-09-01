FROM goalert/build-env:go1.16.4-postgres13 AS build
COPY / /build/
WORKDIR /build
RUN make bin/build/goalert-linux-amd64

FROM alpine
RUN apk --no-cache add ca-certificates
ENV GOALERT_LISTEN :8081
EXPOSE 8081
CMD ["/usr/bin/goalert"]

COPY --from=build /build/bin/build/goalert-linux-amd64/goalert/bin/* /usr/bin/
RUN /usr/bin/goalert self-test
