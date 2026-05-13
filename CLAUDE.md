# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

- **Start server**: `go run backend/cmd/server/main.go` (loads `config.yaml`)
- **Run all tests**: `go test ./...`
- **Run single test**: `go test ./agents -run TestRunner`
- **Test deps**: tests require running MySQL, MongoDB, and Redis instances

## Architecture

### NovelFlow — AI novel-writing platform

Go 1.26 project using CloudWeGo Eino ADK (`deep` mode) + Gin.

### Backend (Gin HTTP)

Standard three-layer: `handler → service → repository`.

| Layer | Path |
|---|---|
| Routes | `backend/internal/route.go` |
| Handlers | `backend/internal/handler/` — parse HTTP, call service |
| Services | `backend/internal/service/` — business logic |
| Middleware | `backend/internal/middleware/` — JWT auth + CORS |
| Response | `backend/internal/response/` — uniform JSON envelope |
| JWT | `backend/pkg/jwt/` — access + refresh tokens, HS256 |
| DI | `backend/internal/servicecontext/svc.go` — wires all deps |

### AI Agent Engine

- **Agent Runner** (`agents/agent.go`): Eino ADK deep agent. Streams output via `RunA()` with thinking/content/tool message types. Configurable dual-model (main LLM + lite LLM for summarization). Retries on 429. Wraps tool errors (never crashes).
- **Main Agent** (`agents/mainagent.go`): novel-writing agent. Loads 12 `.skills/` writing modules as Eino skills with a review sub-agent (novel_review_agent). Provides `write_novel_chapter_file_tool`.
- **System Prompt** (`agents/prompt.go`): Chinese web novel writing rules — scene structure, 4-element chapter design, post-chapter consistency review.
- **Skills** (`.skills/`): 10 Eino-skill modules (concept-planning, opening, volume-outline, plot-logic, character-consistency, transition, dialogue, chapter-ending, anti-ai-voice, consistency-review).

### Data Flow

```
HTTP → Gin → Handler → Service → Repository (MySQL/Redis)
                          ↓ (optional)
                     Agent Runner (Eino deep agent → LLM → file output)
```

### Configuration

- File: `config.yaml` via Viper
- All keys overridable via `NOVELFLOW_*` env vars (bound in `config/config.go`)
- Two LLM configs: `llm` (main agent) and `lite_llm` (summarization)
- Model types: `anthropic` (Claude SDK) or `openai` (OpenAI-compatible)

### Database

- **MySQL** (GORM): User accounts (`users` table, auto-migrated)
- **MongoDB**: Agent chat sessions + messages (collections: `sessions`, `messages`)
- **Redis**: JWT blacklist (logout revocation)

### Workflow

- **README sync**: When making important changes (new features, structural changes, config changes), update `README.md` to reflect them.
- **Commit style**: One commit per logical step, clear Chinese commit messages.

### Key Patterns

- Agent tools must be registered in `agents/tools.go` → `loadAgentTools()`
- Tool errors never abort the conversation — `safeToolMiddleware` converts to string
- Unknown tool calls get a no-op handler that tells the model to retry
- Session IDs are UUIDs; pass `""` to create new, existing UUID to resume
- Test files use `config.LoadConfig("../config.yaml")` — run from module root
