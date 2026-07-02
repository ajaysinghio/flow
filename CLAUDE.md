# flow — CLAUDE.md

## What this is

A neurodivergent-aware task + mood CLI with MCP and REST API interfaces. Single Go binary, local SQLite. No cloud, no auth required locally.

## Architecture

```
internal/
├── app/        ← core aggregate (*app.App) — all presentation layers import this, nothing else
├── store/      ← SQLite via modernc.org/sqlite (pure Go, no CGO)
├── task/       ← Task model, service, engine (Suggest/Ranked), due date parser
├── mood/       ← CheckIn model and service
├── journal/    ← Note capture
├── focus/      ← Focus session timer
├── cli/        ← cobra commands (presentation)
├── api/        ← REST API + OpenAPI 3.1 spec (presentation)
├── mcp/        ← MCP server for Claude (presentation)
└── tray/       ← macOS menu bar / Windows systray (presentation)
```

**Key invariant:** Every presentation layer takes `*app.App`. None of them import `store` or create services directly. The engine (`task.Suggest`, `task.Ranked`) is pure logic — no DB access.

## Running locally

```bash
go run ./cmd/flow          # CLI
go run ./cmd/flow serve    # REST API at localhost:7777
go run ./cmd/flow mcp      # MCP server on stdio
go run ./cmd/flow tray     # macOS menu bar
go run ./cmd/flow pick     # interactive task picker
```

## Install

```bash
go install ./cmd/flow
```

Binary lands at `~/go/bin/flow`. Data at `~/.flow/flow.db`.

## Key design decisions

- **`flow` (no args) returns exactly one task** — never a list. Choice paralysis is the enemy.
- **Energy gates tasks** — `energyThreshold(currentEnergy)` filters out tasks requiring more energy than available. Low energy (1–2) → only low tasks. No pushing through mismatches.
- **`flow pick`** — the "choose and focus" flow: ranked list → select → focus timer → mark done prompt.
- **Due dates boost scoring**: overdue +8, due today +5, due this week +2.
- **Schema migrations** are idempotent `ALTER TABLE … ADD COLUMN` calls in `store.Open()` — safe to run on every startup.
- **No CGO** except for `fyne.io/systray` (macOS/Windows tray only). Everything else is pure Go.

## Database

SQLite at `~/.flow/flow.db`. Tables: `tasks`, `checkins`, `notes`, `focus_sessions`, `migrations`.

Time values are stored as RFC3339 strings by the modernc.org/sqlite driver. The `parseTime` helper in `task/service.go` handles multiple formats.

## MCP tools

`get_context`, `add_task`, `suggest_task`, `checkin`, `breakdown_task`, `complete_task`, `add_note`, `get_insights`

## REST API

OpenAPI spec served at `GET /openapi.json`. Endpoints mirror MCP tools. Auth via `Authorization: Bearer <key>` (optional — set `FLOW_API_KEY` env var or `--api-key` flag).

## Adding a new command

1. Create `internal/cli/<name>.go` with `func newXCmd(a *app.App) *cobra.Command`
2. Register it in `internal/cli/root.go` → `root.AddCommand(newXCmd(a))`
3. Use `a.Tasks`, `a.Moods`, `a.Journal`, `a.Focus` — never `a.DB` directly unless adding a new table

## Adding a new MCP tool

Add to `internal/mcp/tools.go` via `srv.AddTool(...)`. Use `req.GetString(key, default)` and `req.GetInt(key, default)` — never index `req.Params.Arguments` directly.
