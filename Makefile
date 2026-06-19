# Variables
BINARY_NAME=skan
BUILD_DIR=bin
VERSION=1.0.0

.PHONY: all build run clean test help

# Default target runs when you just type 'make'
all: clean build test

## build: Compiles the binary to the bin/ directory
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

## run: Compiles and immediately executes the binary
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

## test: Runs all tests in the project recursively
test:
	@echo "Running tests..."
	@go test -v ./...

## clean: Removes compiled binaries and build artifacts
clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

## help: Shows available commands with descriptions
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'