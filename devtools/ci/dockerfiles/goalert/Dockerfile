FROM alpine

RUN apk --no-cache add tzdata ca-certificates
ENV GOALERT_LISTEN :8081
CMD ["/usr/bin/goalert"]

COPY bin/goalert /usr/bin/goalert
