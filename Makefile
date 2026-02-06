BINARY_NAME := aisk
BUILD_DIR := bin
VERSION := $(shell grep 'AppVersion' internal/config/config.go | head -1 | cut -d'"' -f2)
LDFLAGS := -s -w

.PHONY: build install test lint clean

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aisk

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

test:
	go test ./... -count=1 -race

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

snapshot:
	goreleaser release --snapshot --clean

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test
