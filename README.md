# ShellGate

Turn locally authenticated CLI tools into an OpenAI-compatible REST API.

Log in to a CLI tool once. ShellGate proxies any HTTP client through it — no separate API keys, no extra billing accounts.

```
Your App  →  POST /v1/chat/completions  →  ShellGate  →  codex exec "..."
(N8N, LangChain, any OpenAI SDK)              ↑
                                        your CLI login
```

## Supported Providers

| Provider | CLI | Status |
|----------|-----|--------|
| [Codex](https://github.com/openai/codex) | `codex` | ✅ Supported |
| Claude CLI | `claude` | 🔜 Planned |
| Others | — | 🔜 Planned |

## Install

**One-line (Linux/macOS):**
```bash
curl -fsSL https://raw.githubusercontent.com/DutaKey/ShellGate/main/install.sh | sh
```

**Homebrew:** *(coming soon)*
```bash
brew install dutakey/tap/shellgate
```

**From source:**
```bash
git clone https://github.com/DutaKey/ShellGate
cd ShellGate && make build
```

## Quick Start

```bash
# 1. Create config
shellgate init

# 2. Log in to your CLI provider
shellgate login codex

# 3. Start the API server
shellgate serve

# 4. Create an API key for your project
shellgate keys create myproject
```

Done. Use `http://localhost:8080/v1` as your OpenAI base URL.

## CLI Reference

```
shellgate init                  Interactive setup wizard
shellgate login <provider>      Authenticate a CLI provider
shellgate serve                 Start the API server
shellgate keys create <name>    Create a new API key
shellgate keys list             List all keys
shellgate keys revoke <id>      Revoke a key
```

**Flags:**
```
-c, --config string   config file path (default "config.toml")
```

## API

ShellGate is a drop-in replacement for the OpenAI API.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/chat/completions` | Chat completions (streaming + non-streaming) |
| `POST` | `/v1/responses` | Responses API (N8N AI Agent, LangChain) |
| `GET` | `/v1/models` | List available models |
| `GET` | `/v1/models/:id` | Get model by ID |
| `GET` | `/health` | Health check |
| `POST` | `/admin/keys` | Create API key |
| `GET` | `/admin/keys` | List API keys |
| `DELETE` | `/admin/keys/:id` | Revoke API key |

## Usage Examples

**curl:**
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <your-key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.4","messages":[{"role":"user","content":"hello"}]}'
```

**Python (OpenAI SDK):**
```python
from openai import OpenAI

client = OpenAI(
    api_key="<your-shellgate-key>",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-5.4",
    messages=[{"role": "user", "content": "hello"}]
)
```

**N8N:** OpenAI node or AI Agent node → set Base URL to `http://<host>:8080/v1`.

## Configuration

```toml
[server]
host = "0.0.0.0"
port = 8080
read_timeout = "30s"
write_timeout = "120s"

[auth]
admin_secret = ""        # required — protects /admin/* endpoints
keys_file = "keys.json"

[executor]
codex_binary = "codex"
default_sandbox = "read-only"
timeout = "120s"

[logging]
level = "info"   # debug | info | warn | error
format = "json"  # json | text
```

Environment variable overrides:

| Variable | Config field |
|----------|-------------|
| `SHELLGATE_PORT` | `server.port` |
| `SHELLGATE_ADMIN_SECRET` | `auth.admin_secret` |
| `SHELLGATE_KEYS_FILE` | `auth.keys_file` |
| `SHELLGATE_EXECUTOR_CODEX_BINARY` | `executor.codex_binary` |
| `SHELLGATE_LOG_LEVEL` | `logging.level` |

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $HOME/.codex:/root/.codex:ro \
  -e SHELLGATE_ADMIN_SECRET=your-secret \
  -e SHELLGATE_KEYS_FILE=/app/data/keys.json \
  ghcr.io/dutakey/shellgate:latest
```

## License

MIT
