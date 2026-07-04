BINARY     := tview
CMD        := ./cmd/tview
INSTALL_DIR ?= $(HOME)/.local/bin
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE       := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    := -s -w \
              -X main.version=$(VERSION) \
              -X main.commit=$(COMMIT) \
              -X main.date=$(DATE)

.PHONY: all build run install-local clean lint tidy test snapshot

## build: Compile the binary for the current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

## install-local: Build and install tview to ~/.local/bin
install-local: build
	mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"

## run: Build and run with example args (edit as needed)
run: build
	./$(BINARY) btc 15m

## tidy: Tidy go modules
tidy:
	go mod tidy

## lint: Run golangci-lint (install separately: https://golangci-lint.run)
lint:
	golangci-lint run ./...

## test: Run unit tests
test:
	go test ./... -v -race

## snapshot: Build a local snapshot release (requires goreleaser)
snapshot:
	goreleaser release --snapshot --clean

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
