FROM alpine:3.18.3

WORKDIR /app

COPY bin/queue-share /app/queue-share

CMD ["/app/queue-share"]
