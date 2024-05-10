FROM arm64v8/alpine

WORKDIR /app
RUN mkdir -p /app/temp

COPY bin/queue-share /app/queue-share
CMD ["/app/queue-share"]
