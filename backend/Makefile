.PHONY: all test swag mock build run clean

all: test build

# Run unit tests
test:
	go test -v ./...

# Generate Swagger documentation
swag:
	swag init -g cmd/api/main.go -o api/swagger

# Generate mocks
mock:
	go generate ./...

# Build API binary
build:
	go build -o bin/api cmd/api/main.go

# Run API server
run:
	go run cmd/api/main.go

# Clean build artifacts
clean:
	rm -rf bin/ api/swagger/
