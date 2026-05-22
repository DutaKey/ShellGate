FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shellgate ./cmd/shellgate


FROM node:22-alpine AS runtime

# Install codex CLI
RUN npm install -g @openai/codex

# Copy ShellGate binary
COPY --from=builder /shellgate /usr/local/bin/shellgate

WORKDIR /app

EXPOSE 8080

# keys.json persisted via volume mount: -v /host/data:/app/data
# Set SHELLGATE_ADMIN_SECRET and SHELLGATE_KEYS_FILE=/app/data/keys.json
ENTRYPOINT ["shellgate", "-config", "/app/config.toml"]
