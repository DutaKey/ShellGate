# ShellGate

OpenAI-compatible API proxy for the [Codex CLI](https://github.com/openai/codex). Wrap your local CLI session into a drop-in REST API — no extra API keys required.

## How it works

```
Your App (OpenAI SDK)
       │
       │  POST /v1/chat/completions
       │  Authorization: Bearer sk-sg-xxxx
       ▼
   ShellGate  ──────────────────────────────────────────┐
       │                                                 │
       │  codex exec --json "<prompt>"                   │
       ▼                                                 │
  Codex CLI (saved OAuth session)           Admin API    │
       │                                  /admin/keys    │
       │  JSONL events                                   │
       ▼                                                 │
  OpenAI-compatible response ◄─────────────────────────┘
```

One Codex account. Multiple projects. Each project gets its own ShellGate API key.

## Requirements

- Go 1.21+
- [Codex CLI](https://github.com/openai/codex) installed and authenticated (`codex login`)

## Installation

**From source:**
```bash
git clone https://github.com/dutakey/shellgate
cd shellgate
make build
```

**Docker:**
```bash
docker build -t shellgate .
```

## Setup

**1. Authenticate Codex CLI on your server:**
```bash
codex login
```

**2. Create config:**
```bash
cp config.example.toml config.toml
```

Edit `config.toml`:
```toml
[auth]
admin_secret = "your-strong-admin-secret"
```

**3. Start ShellGate:**
```bash
./bin/shellgate -config config.toml
# or
make run
```

**4. Create API keys for your projects:**
```bash
# Create key for N8N
curl -X POST http://localhost:8080/admin/keys \
  -H "Authorization: Bearer your-strong-admin-secret" \
  -H "Content-Type: application/json" \
  -d '{"name": "n8n"}'

# Create key for Laravel
curl -X POST http://localhost:8080/admin/keys \
  -H "Authorization: Bearer your-strong-admin-secret" \
  -H "Content-Type: application/json" \
  -d '{"name": "laravel"}'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "key": "sk-sg-a1b2c3d4...",
  "name": "n8n",
  "created_at": "2026-01-01T00:00:00Z"
}
```

## Usage

ShellGate is a drop-in replacement for the OpenAI API. Point any OpenAI-compatible client to your ShellGate instance.

**curl:**
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-sg-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "codex",
    "messages": [
      {"role": "user", "content": "Write a Go hello world"}
    ]
  }'
```

**Python (OpenAI SDK):**
```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-sg-your-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="codex",
    messages=[{"role": "user", "content": "Write a Go hello world"}]
)
```

**Streaming:**
```python
stream = client.chat.completions.create(
    model="codex",
    messages=[{"role": "user", "content": "Explain Go channels"}],
    stream=True
)
for chunk in stream:
    print(chunk.choices[0].delta.content or "", end="")
```

**N8N:** Use the OpenAI node, set `Base URL` to `http://your-server:8080/v1`.

## API Reference

### OpenAI-Compatible Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/chat/completions` | Chat completions (streaming + non-streaming) |
| `GET` | `/v1/models` | List available models |
| `GET` | `/health` | Health check |

### Admin Endpoints

All admin endpoints require `Authorization: Bearer <admin_secret>`.

| Method | Path | Body | Description |
|--------|------|------|-------------|
| `POST` | `/admin/keys` | `{"name": "string"}` | Create API key |
| `GET` | `/admin/keys` | — | List all keys |
| `DELETE` | `/admin/keys/:id` | — | Revoke key by ID |

## Configuration

```toml
[server]
host = "0.0.0.0"
port = 8080
read_timeout = "30s"
write_timeout = "120s"     # increase for long-running codex tasks

[auth]
admin_secret = ""          # required — protects /admin/* endpoints
keys_file = "keys.json"    # where API keys are stored

[executor]
codex_binary = "codex"     # full path if not in $PATH
default_sandbox = "read-only"   # read-only | workspace-write | danger-full-access
timeout = "120s"
working_dir = ""           # default: inherit ShellGate cwd

[logging]
level = "info"             # debug | info | warn | error
format = "json"            # json | text
```

All fields can be overridden via environment variables:

| Variable | Config field |
|----------|-------------|
| `SHELLGATE_PORT` | `server.port` |
| `SHELLGATE_HOST` | `server.host` |
| `SHELLGATE_ADMIN_SECRET` | `auth.admin_secret` |
| `SHELLGATE_KEYS_FILE` | `auth.keys_file` |
| `SHELLGATE_EXECUTOR_CODEX_BINARY` | `executor.codex_binary` |
| `SHELLGATE_EXECUTOR_SANDBOX` | `executor.default_sandbox` |
| `SHELLGATE_LOG_LEVEL` | `logging.level` |

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e SHELLGATE_ADMIN_SECRET=your-secret \
  -e SHELLGATE_KEYS_FILE=/app/data/keys.json \
  shellgate
```

Mount `/app/data` to persist keys across container restarts.

> **Note:** Codex CLI must be authenticated. Mount your `~/.codex` credentials into the container:
> ```bash
> -v $HOME/.codex:/root/.codex:ro
> ```

## Notes

- `usage` tokens in responses are always `-1` — Codex CLI does not expose token counts
- Each request spawns a new `codex exec` process (stateless)
- Conversation history is handled by concatenating `messages[]` into a single prompt

## License

MIT
