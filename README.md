# ✦ CLAI

**A high-performance terminal chat client for LLMs.**
Built in Go with the [Charm](https://charm.sh) ecosystem — single binary, instant startup, beautiful UI.

---

## Why CLAI?

Most terminal LLM tools fall into two camps: bloated agent frameworks with hundreds of thousands of lines of code, or bare-bones scripts with no UI to speak of. CLAI is neither.

It does one thing — chat with LLMs — and does it exceptionally well. Real-time streaming, full Markdown rendering, syntax-highlighted code blocks, multi-session management, and an in-terminal settings panel. All in a self-contained binary that starts in under 50ms.

---

## Features

- **Real-time streaming** — tokens appear as they're generated, with live Markdown re-rendering
- **Syntax highlighting** — 100+ languages via Chroma, embedded in the binary
- **Multi-session management** — create, switch, rename, and delete conversations
- **In-TUI settings** — configure model, temperature, system prompt, and API endpoint without leaving the terminal
- **OpenAI-compatible** — works with OpenAI, Groq, OpenRouter, Together.ai, Fireworks, Ollama, or any OpenAI-compatible server
- **Adaptive theme** — automatic dark/light mode based on terminal background
- **Zero dependencies at runtime** — single static binary, no Node.js, no Python, no Docker
- **Persistent sessions** — conversations saved as plain JSON in `~/.local/share/clai/`

---

## Installation

### From source

Requires Go 1.22+.

```bash
git clone https://github.com/lioarce01/clai.git
cd clai
go build -o clai ./cmd/clai
```

Move to your PATH:

```bash
mv clai /usr/local/bin/
```

### go install

```bash
go install github.com/lioarce01/clai/cmd/clai@latest
```

---

## Quick Start

```bash
export OPENAI_API_KEY=sk-...
clai
```

That's it. A default session is created automatically.

---

## Configuration

CLAI is configured via environment variables, CLI flags, or the in-app settings panel (`Ctrl+O`). Settings are saved to `~/.config/clai/config.toml`.

### Environment variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | API key for your provider |
| `OPENAI_BASE_URL` | Base URL override (default: `https://api.openai.com/v1`) |

### CLI flags

```
Usage: clai [flags]

  --api-key    string   API key (overrides OPENAI_API_KEY)
  --base-url   string   Base URL for the API endpoint
  --model      string   Model to use (default: gpt-4o)
  --help                Show help
```

### Config file

`~/.config/clai/config.toml` is created on first launch and editable directly:

```toml
[api]
api_key  = ""                          # or set OPENAI_API_KEY
base_url = "https://api.openai.com/v1"

[model]
name          = "gpt-4o"
temperature   = 0.7
max_tokens    = 4096
top_p         = 1.0
system_prompt = "You are a helpful assistant."
```

---

## Compatible Providers

Any OpenAI-compatible endpoint works by setting `--base-url`:

| Provider | Base URL | Notes |
|----------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | Default |
| Groq | `https://api.groq.com/openai/v1` | Free tier, very fast |
| OpenRouter | `https://openrouter.ai/api/v1` | 100+ models |
| Together.ai | `https://api.together.xyz/v1` | Open-source models |
| Fireworks | `https://api.fireworks.ai/inference/v1` | |
| Ollama | `http://localhost:11434/v1` | Local models, set `--api-key ollama` |

---

## Key Bindings

### Chat

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | Insert new line |
| `↑ / ↓` | Scroll message history |
| `PgUp / PgDn` | Page through history |
| `Home / End` | Jump to top / bottom |

### Global

| Key | Action |
|-----|--------|
| `Ctrl+N` | New session |
| `Ctrl+S` | Open session picker |
| `Ctrl+O` | Open settings panel |
| `Ctrl+L` | Clear chat view |
| `Ctrl+C` | Quit |

### Overlays

| Key | Action |
|-----|--------|
| `Esc` | Close overlay |
| `Tab / Shift+Tab` | Next / previous field (settings) |
| `Enter` | Select / save |
| `n` | New session (session picker) |
| `d` | Delete session (session picker) |
| `/` | Filter sessions (session picker) |

---

## Architecture

```
clai/
├── cmd/clai/            # Entry point
├── internal/
│   ├── config/          # TOML config, env overrides, defaults
│   ├── llm/             # LLM client interface + OpenAI implementation
│   ├── storage/         # Session persistence (JSON files)
│   ├── markdown/        # Glamour renderer wrapper
│   └── tui/             # Bubble Tea UI components
```

**Design principles:**

- **Interface-driven** — `llm.Client` and `storage.Store` are interfaces; swap providers without touching the UI
- **Elm architecture** — every TUI component follows Bubble Tea's `Model → Update → View` pattern
- **Streaming-first** — the LLM client returns `<-chan StreamDelta`; the UI subscribes and re-renders incrementally
- **Zero global state** — all state lives in the Bubble Tea model tree; no singletons, no `init()` side effects

---

## Performance

| Metric | CLAI | OpenCode | OpenClaw |
|--------|------|----------|----------|
| Cold start | ~30ms | ~200ms | ~2–5s |
| Memory (idle) | ~20MB | ~80MB | ~150MB |
| Binary size | ~15MB | ~25MB | N/A (Node.js) |
| Dependencies | 0 runtime | 0 runtime | Node.js + npm |

---

## Roadmap

| Version | Feature |
|---------|---------|
| **v0.1.0** | MVP — Chat, streaming, sessions, settings ✓ |
| v0.2.0 | Native Anthropic API support |
| v0.3.0 | Google Gemini support |
| v0.4.0 | Ollama / local models |
| v0.5.0 | Export conversations (Markdown, JSON, clipboard) |
| v0.6.0 | Image input (GPT-4o vision) |
| v0.7.0 | Custom themes |
| v1.0.0 | Stable release |

---

## License

MIT — see [LICENSE](LICENSE).
