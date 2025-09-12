FROM docker.io/goalert/build-env:go1.25.0 AS build

# Ensure we can write to the build workspace (base image runs as a non-root user)
USER root
COPY . /build/
RUN chown -R user:user /build
USER user
WORKDIR /build
RUN make clean bin/build/goalert-linux-amd64 || ( echo "Make failed; printing env for diagnostics" && env && exit 1 )

FROM docker.io/library/alpine
RUN apk --no-cache add ca-certificates
ENV GOALERT_LISTEN=:8081
EXPOSE 8081
CMD ["/usr/bin/goalert"]

COPY --from=build /build/bin/build/goalert-linux-amd64/goalert/bin/* /usr/bin/
RUN /usr/bin/goalert self-test