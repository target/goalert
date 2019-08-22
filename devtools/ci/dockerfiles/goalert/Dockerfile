FROM alpine

RUN apk --no-cache add tzdata ca-certificates
ENV GOALERT_LISTEN :8081

COPY bin/goalert /usr/bin/goalert
