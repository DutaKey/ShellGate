BINARY := shellgate
BUILD_DIR := bin
GO := $(shell which go 2>/dev/null || echo /home/dxtz/go/bin/go)
GOFLAGS := -ldflags="-s -w"
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build run dev clean test lint release

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/shellgate

run: build
	./$(BUILD_DIR)/$(BINARY) serve

dev:
	$(GO) run ./cmd/shellgate serve

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

release:
	GOOS=linux   GOARCH=amd64  $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY)_linux_amd64   ./cmd/shellgate
	GOOS=linux   GOARCH=arm64  $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY)_linux_arm64   ./cmd/shellgate
	GOOS=darwin  GOARCH=amd64  $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY)_darwin_amd64  ./cmd/shellgate
	GOOS=darwin  GOARCH=arm64  $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY)_darwin_arm64  ./cmd/shellgate
	GOOS=windows GOARCH=amd64  $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY)_windows_amd64.exe ./cmd/shellgate

docker-build:
	docker build -t shellgate:latest .

docker-run:
	docker run -p 8080:8080 \
		-e SHELLGATE_ADMIN_SECRET=your-secret \
		-e SHELLGATE_KEYS_FILE=/app/data/keys.json \
		shellgate:latest
