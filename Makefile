.PHONY: all build clean test lint examples benchmark simple align generate install

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOLINT=golangci-lint

# Binary name
BINARY_NAME=quickvalidate
BINARY_PATH=./cmd/quickvalidate

# Main build target
all: build

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Run tests
test:
	$(GOTEST) -v ./...

# Run linting
lint:
	$(GOLINT) run

# Run all examples
examples: simple benchmark

# Run simple example
simple:
	$(GOCMD) run ./examples/simple

# Run benchmark example
benchmark:
	$(GOCMD) run ./examples/benchmark

# Run the differential alignment suite: generated code vs go-playground/validator
align: build
	$(GOTEST) -count=1 ./internal/align/

# Generate validation code for every package that needs it
generate: build
	./$(BINARY_NAME) -i ./examples/simple -o ./examples/simple/validate_gen.go
	./$(BINARY_NAME) -i ./examples/numeric -o ./examples/numeric/validate_gen.go
	./$(BINARY_NAME) -i ./examples/benchmark -o ./examples/benchmark/validate_gen.go
	./$(BINARY_NAME) -i ./internal/align -o ./internal/align/validate_gen.go -pkg align

# Install the binary
install:
	$(GOINSTALL) $(BINARY_PATH)

# Update dependencies
deps:
	$(GOCMD) mod tidy
	$(GOCMD) mod verify

# Help target
help:
	@echo "Available targets:"
	@echo "  all        - Build the binary"
	@echo "  build      - Build the binary"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linting"
	@echo "  examples   - Run all examples"
	@echo "  simple     - Run simple example"
	@echo "  benchmark  - Run benchmark example"
	@echo "  align      - Run the differential alignment suite"
	@echo "  generate   - Regenerate validation code for examples and corpus"
	@echo "  install    - Install the binary"
	@echo "  deps       - Update dependencies"
	@echo "  help       - Show this help message"
