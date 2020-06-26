FROM scratch AS extract
ADD bin/goalert-linux-amd64.tgz /files/

FROM alpine
RUN apk --no-cache add tzdata ca-certificates
ENV GOALERT_LISTEN :8081
EXPOSE 8081
CMD ["/usr/bin/goalert"]

COPY --from=extract /files/goalert/bin/goalert /usr/bin/goalert
