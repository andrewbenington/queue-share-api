FROM amd64/alpine:3.14

WORKDIR /app
RUN mkdir -p /app/temp

LABEL fix="force-reupload-2025-06-12"

COPY bin/queue-share /app/queue-share
CMD ["/app/queue-share"]
