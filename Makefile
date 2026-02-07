.PHONY: run clean build dev lint

# Default target
run: build
	@echo "Starting botTaskTracker..."
	@/usr/local/go/bin/go run .

# Build assets and generate code
build:
	@echo "Building CSS assets..."
	@npm run build:linux
	@echo "Generating templ files..."
	@/usr/local/go/bin/go generate ./...

# Lint code
lint:
	@echo "Running golangci-lint..."
	@go tool golangci-lint run --timeout 5m

# Clean up binaries and build artifacts
clean:
	@echo "Cleaning up binaries..."
	@rm -f botTaskTracker botTaskTraacker
	@echo "Cleaned."

# Development mode with hot reload (if needed)
dev:
	@echo "Starting development mode..."
	@npm run build:linux
	@/usr/local/go/bin/go generate ./...
	@/usr/local/go/bin/go run .
