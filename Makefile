.PHONY: start
start:
	@go run ./cmd/main.go

.PHONY: build
build:
	go build -o bin/go-spotify ./cmd/main.go