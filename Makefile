BIN=gin-template
PKG=github.com/wiidz/gin_template/cmd/server

.PHONY: run build tidy fmt lint test dev

run:
	go run ./cmd/server

build:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BIN) ./cmd/server

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

dev:
	command -v air >/dev/null 2>&1 && air -c .air.toml || go run github.com/air-verse/air@latest -c .air.toml


