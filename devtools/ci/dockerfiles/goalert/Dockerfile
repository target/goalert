FROM alpine
RUN apk --no-cache add tzdata ca-certificates
COPY goalert /usr/bin/
CMD ["/usr/bin/goalert"]
