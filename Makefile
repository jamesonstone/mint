.PHONY: build build-windows test lint fmt vet clean install install-git-hooks tidy all

BINARY_NAME=mint
VERSION?=$(shell git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-X github.com/jamesonstone/mint/pkg/cli.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/mint

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME).exe ./cmd/mint

install:
	go install $(LDFLAGS) ./cmd/mint

install-git-hooks:
	chmod +x .githooks/pre-commit
	git config core.hooksPath .githooks

test:
	go test -v ./...
	python3 -m unittest discover -s adapters/tests -v

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	go clean

tidy:
	go mod tidy

all: fmt vet test build

# Activate .venv-docs first; see website/README.md.
.PHONY: docs-build docs-check docs-serve
docs-build:
	python3 scripts/build_docs.py

docs-check:
	python3 -m unittest discover -s scripts -p 'test_*.py'
	python3 scripts/build_docs.py

docs-serve: docs-build
	python3 -m http.server 8000 --bind 127.0.0.1 --directory _site
