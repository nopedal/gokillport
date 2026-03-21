BINARY    := gokillport
MODULE    := github.com/nikolasopedal/gokillport
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
BUILD_DIR := dist

.PHONY: build install clean cross test

## build: compile for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/gokillport

## install: install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gokillport

## test: run all tests
test:
	go test ./...

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) $(BUILD_DIR)

## cross: build for all platforms
cross: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64   ./cmd/gokillport
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64   ./cmd/gokillport
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64    ./cmd/gokillport
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64    ./cmd/gokillport
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/gokillport
	@echo "Built binaries in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/
