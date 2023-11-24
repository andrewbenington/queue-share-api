FROM arm64v8/alpine

WORKDIR /app

COPY bin/queue-share /app/queue-share
CMD ["/app/queue-share"]
