.PHONY: build run dev test lint clean migrate-up migrate-down docker-build

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run: build
	./bin/server

# Run in development mode with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations (requires golang-migrate)
# Default database URL for local development
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/picklego?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-down-all:
	migrate -path migrations -database "$(DATABASE_URL)" down -all

migrate-status:
	migrate -path migrations -database "$(DATABASE_URL)" version

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path migrations -database "$(DATABASE_URL)" force $$version

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-reset: migrate-down-all migrate-up
	@echo "Database reset complete"

# Docker
docker-build:
	docker build -t pickle-go-api .

docker-run:
	docker run -p 8080:8080 --env-file .env pickle-go-api

# Dependencies
deps:
	go mod download
	go mod tidy

# Install development tools
tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate OpenAPI client
openapi:
	oapi-codegen -package api -generate types,client api/openapi.yaml > api/client.gen.go

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Build and run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make deps           - Download dependencies"
	@echo "  make tools          - Install development tools"
