# Makefile
.PHONY: build run test clean dev build-app

# Development: start backend and frontend dev servers
dev:
	@echo "Starting backend and frontend development servers..."
	@go run cmd/server/main.go &
	@cd web && npm run dev

# Build app: build backend, build frontend, and run preview
build-app: build
	@echo "Building frontend..."
	@cd web && npm run build
	@echo "Starting preview server..."
	@cd web && npm run preview

# Run the application
run: build
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/
	go clean

# Development setup
dev-setup:
	go mod tidy
	go install github.com/air-verse/air@latest
	cd web && npm install

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

migrate:
	go run ./cmd/migrate

# Install dependencies
deps:
	cd web && npm install