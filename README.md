# ShellGate

Turn locally authenticated CLI tools into an OpenAI-compatible REST API.

Log in to a CLI tool once. ShellGate proxies any HTTP client through it — no separate API keys, no extra billing accounts.

```
Your App  →  POST /v1/chat/completions  →  ShellGate  →  codex exec "..."
(N8N, LangChain, any OpenAI SDK)                      →  kimi --print "..."
                                              ↑
                                        your CLI logins
```

Requests route automatically based on model name — use Codex and Kimi simultaneously from the same endpoint.

## Supported Providers

| Provider | CLI | Model prefix | Status |
|----------|-----|-------------|--------|
| [OpenAI Codex](https://github.com/openai/codex) | `codex` | `gpt-*` | ✅ Supported |
| [Kimi CLI](https://moonshotai.github.io/kimi-cli/) | `kimi` | `kimi-*` | ✅ Supported |
| Antigravity CLI | `agy` | — | 🔜 Planned |
| Claude CLI | `claude` | — | 🔜 Planned |

## Install

**One-line (Linux/macOS):**
```bash
curl -fsSL https://raw.githubusercontent.com/DutaKey/ShellGate/main/install.sh | sh
```

**From source:**
```bash
git clone https://github.com/DutaKey/ShellGate
cd ShellGate && make build
```

## Quick Start

### Option A — Guided setup wizard
```bash
shellgate setup
```

Walks through config creation, provider login, API key generation, and server start in one flow.

### Option B — Manual
```bash
# 1. Create config (~/.shellgate/config.toml)
shellgate init

# 2. Log in to CLI providers
shellgate login codex
shellgate login kimi

# 3. Start the server
shellgate serve -d          # background
# or
shellgate serve             # foreground

# 4. Create an API key
shellgate keys
```

Use `http://localhost:8080/v1` as your OpenAI base URL.

## Model Routing

ShellGate routes each request to the right provider based on model name:

| Model | Provider |
|-------|----------|
| `gpt-5.5`, `gpt-5.4`, `gpt-5.4-mini`, `gpt-5.3-codex`, `gpt-5.2` | Codex CLI |
| `kimi-code/kimi-for-coding` | Kimi CLI |

Both providers are always active — no config switch needed.

## CLI Reference

```
shellgate setup                 Guided first-time setup wizard
shellgate init                  Create config file interactively
shellgate login <provider>      Authenticate a CLI provider (shows current auth status)
shellgate serve                 Start the API server (foreground)
shellgate serve -d              Start in background
shellgate stop                  Stop background server
shellgate restart               Restart background server
shellgate status                Show server status + provider auth status
shellgate logs                  View last 50 log lines
shellgate logs -f               Follow log output (tail -f style)
shellgate keys                  Interactive key manager (arrow keys)
shellgate keys create <name>    Create API key (scriptable)
shellgate keys list             List all keys
shellgate keys revoke <id>      Revoke a key
```

**Flags:**
```
-c, --config string   config file (default ~/.shellgate/config.toml)
```

## API

Drop-in replacement for the OpenAI API.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/chat/completions` | Chat completions (streaming + non-streaming) |
| `POST` | `/v1/responses` | Responses API (N8N AI Agent, LangChain) |
| `GET` | `/v1/models` | List available models (all providers) |
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

**Kimi model:**
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <your-key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"kimi-code/kimi-for-coding","messages":[{"role":"user","content":"hello"}]}'
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

**LangChain / other frameworks:** Any client that supports custom OpenAI base URL works.

## Configuration

Config lives at `~/.shellgate/config.toml` by default.

```toml
[server]
host = "0.0.0.0"
port = 8080
read_timeout = "30s"
write_timeout = "120s"

[auth]
admin_secret = ""                          # required — protects /admin/* endpoints
keys_file = "~/.shellgate/keys.json"

[executor]
codex_binary = "codex"
default_sandbox = "read-only"              # read-only | workspace-write | danger-full-access
timeout = "120s"
kimi_binary = "kimi"

[logging]
level = "info"    # debug | info | warn | error
format = "json"   # json | text
```

**Environment variable overrides:**

| Variable | Config field |
|----------|-------------|
| `SHELLGATE_PORT` | `server.port` |
| `SHELLGATE_ADMIN_SECRET` | `auth.admin_secret` |
| `SHELLGATE_KEYS_FILE` | `auth.keys_file` |
| `SHELLGATE_EXECUTOR_CODEX_BINARY` | `executor.codex_binary` |
| `SHELLGATE_EXECUTOR_KIMI_BINARY` | `executor.kimi_binary` |
| `SHELLGATE_LOG_LEVEL` | `logging.level` |

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v $HOME/.shellgate:/root/.shellgate \
  -v $HOME/.codex:/root/.codex:ro \
  -v $HOME/.kimi:/root/.kimi:ro \
  -e SHELLGATE_ADMIN_SECRET=your-secret \
  ghcr.io/dutakey/shellgate:latest
```

## License

MIT
