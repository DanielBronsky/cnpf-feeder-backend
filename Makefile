.PHONY: generate build run test clean

# Generate GraphQL code
generate:
	go generate ./...

# Build the application
build:
	go build -o bin/main ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f main
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run
