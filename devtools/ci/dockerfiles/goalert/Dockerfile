FROM node:10 AS UI
COPY . /build
RUN cd /build && make web/src/build/index.html

FROM golang:alpine AS BE
COPY --from=UI /build /build
RUN apk --no-cache add git make
RUN cd /build && make bin/goalert BUNDLE=1

FROM alpine
RUN apk --no-cache add tzdata ca-certificates
COPY --from=BE /build/bin/goalert /usr/bin/
CMD ["/usr/bin/goalert"]
ENV GOALERT_LISTEN :8081
