# Changelog

All notable changes to ShellGate will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-05-22

### Added
- OpenAI-compatible REST API (`POST /v1/chat/completions`, `GET /v1/models`)
- Codex CLI proxy via `codex exec --json` — no extra API keys needed
- Streaming responses via Server-Sent Events (SSE)
- Admin API for key management (`POST/GET/DELETE /admin/keys`)
- JSON-based API key store with thread-safe read/write
- TOML config with full environment variable override support
- Graceful shutdown with in-flight request draining
- Structured logging via `zap`
- Dockerfile with multi-stage build (Go builder + Node runtime for Codex CLI)
