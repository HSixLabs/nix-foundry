.PHONY: all build test lint clean

all: lint test build

build:
	go build -o nf ./cmd/nix-foundry

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f nix-foundry
	go clean
	rm -rf .go

fmt:
	go fmt ./...
	nixpkgs-fmt *.nix

pre-commit:
	pre-commit run --all-files
