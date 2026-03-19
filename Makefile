# gostack - Fake OpenStack API for integration tests
SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c

BINARY_NAME ?= fake-openstack
GOFLAGS ?= -v
LDFLAGS ?= -s -w

.PHONY: all build test lint clean help

all: build

build:
	go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/fake-openstack

test:
	go test ./... -v -count=1

lint:
	@which golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed, run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

clean:
	rm -rf bin/
	go clean -cache -testcache

help:
	@echo "Targets:"
	@echo "  build   - Build fake-openstack binary (default)"
	@echo "  test    - Run tests"
	@echo "  lint    - Run golangci-lint"
	@echo "  clean   - Remove build artifacts"
