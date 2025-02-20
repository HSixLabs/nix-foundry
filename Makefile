.PHONY: all build test lint clean fmt pre-commit deps

BINARY_NAME := nix-foundry

all: deps lint test build

build:
	go build -o $(BINARY_NAME) .

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
	go clean
	rm -rf dist/

fmt:
	go fmt ./...
	nixpkgs-fmt *.nix

pre-commit:
	pre-commit run --all-files

deps:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/nix-community/nixpkgs-fmt@latest
