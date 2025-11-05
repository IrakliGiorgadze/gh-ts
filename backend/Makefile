SHELL := /bin/bash

run:
	APP_ENV=dev API_PORT=8080 go run ./cmd/api

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

# DB tools: assume goose installed (or use golang-migrate)
# brew install goose
migrate-up:
	goose -dir ./internal/database/migrations postgres "$$DB_DSN" up

migrate-down:
	goose -dir ./internal/database/migrations postgres "$$DB_DSN" down