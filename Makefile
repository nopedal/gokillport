BINARY    := gokillport
ALIAS     := killport
MODULE    := github.com/nopedal/gokillport
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
BUILD_DIR := dist

.PHONY: build install clean cross test release

## build: compile both binaries for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/gokillport
	go build -ldflags "$(LDFLAGS)" -o $(ALIAS) ./cmd/killport

## install: install both binaries to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gokillport
	go install -ldflags "$(LDFLAGS)" ./cmd/killport

## test: run all tests
test:
	go test ./...

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) $(ALIAS) $(BUILD_DIR)

## cross: build for all platforms
cross: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64       ./cmd/gokillport
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64       ./cmd/gokillport
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64        ./cmd/gokillport
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64        ./cmd/gokillport
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe  ./cmd/gokillport
	@echo "Built binaries in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

## release: create a GitHub release with cross-compiled binaries (requires gh CLI)
release: cross
	@if [ "$(VERSION)" = "dev" ]; then echo "error: tag the commit first (git tag v0.1.0)"; exit 1; fi
	gh release create $(VERSION) $(BUILD_DIR)/* --title "$(VERSION)" --generate-notes
