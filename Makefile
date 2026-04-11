.PHONY: help build run test clean

help:
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  run     - Build and run the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"

build:
	go build -o invoice-generator ./cmd/

run: build
	./invoice-generator

test:
	go test -v ./...

clean:
	rm -f invoice-generator
	go clean
