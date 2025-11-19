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
# Load DB_DSN from .env file in project root
migrate-up:
	@if [ -f ../.env ]; then \
		DB_DSN=$$(grep "^DB_DSN=" ../.env | cut -d'=' -f2- | sed 's/^"//;s/"$$//') && \
		if [ -z "$$DB_DSN" ]; then \
			echo "Error: DB_DSN not found in .env file."; \
			exit 1; \
		fi && \
		goose -dir ./internal/database/migrations postgres "$$DB_DSN" up; \
	else \
		if [ -z "$$DB_DSN" ]; then \
			echo "Error: DB_DSN not set and .env file not found."; \
			echo "Please create .env file in project root or export DB_DSN."; \
			echo "Example: export DB_DSN='postgres://ticketuser:ticketpass123@localhost:5432/ticketing_db?sslmode=disable'"; \
			exit 1; \
		else \
			goose -dir ./internal/database/migrations postgres "$$DB_DSN" up; \
		fi \
	fi

migrate-down:
	@if [ -f ../.env ]; then \
		DB_DSN=$$(grep "^DB_DSN=" ../.env | cut -d'=' -f2- | sed 's/^"//;s/"$$//') && \
		if [ -z "$$DB_DSN" ]; then \
			echo "Error: DB_DSN not found in .env file."; \
			exit 1; \
		fi && \
		goose -dir ./internal/database/migrations postgres "$$DB_DSN" down; \
	else \
		if [ -z "$$DB_DSN" ]; then \
			echo "Error: DB_DSN not set and .env file not found."; \
			echo "Please create .env file in project root or export DB_DSN."; \
			echo "Example: export DB_DSN='postgres://ticketuser:ticketpass123@localhost:5432/ticketing_db?sslmode=disable'"; \
			exit 1; \
		else \
			goose -dir ./internal/database/migrations postgres "$$DB_DSN" down; \
		fi \
	fi