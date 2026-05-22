BINARY := shellgate
BUILD_DIR := bin
GO := go
GOFLAGS := -ldflags="-s -w"

.PHONY: build run dev clean test lint

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/shellgate

run: build
	./$(BUILD_DIR)/$(BINARY) -config config.toml

dev:
	$(GO) run ./cmd/shellgate -config config.toml

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t shellgate:latest .

docker-run:
	docker run -p 8080:8080 \
		-e SHELLGATE_AUTH_API_KEY=sk-shellgate-test \
		-e SHELLGATE_AUTH_CODEX_API_KEY=$$CODEX_API_KEY \
		shellgate:latest
