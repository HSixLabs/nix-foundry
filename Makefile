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
	@if [ -z "$$NIXPKGS_FMT" ]; then \
		echo "Error: nixpkgs-fmt not found. Please enter the Nix shell first:"; \
		echo "  nix-shell"; \
		echo "Then run:"; \
		echo "  make fmt"; \
		exit 1; \
	fi
	go fmt ./...
	npx prettier --write "**/*.{yml,yaml,md,json}"
	$(NIXPKGS_FMT) $$(find . -name '*.nix')

pre-commit:
	pre-commit run --all-files

deps:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	npm install
