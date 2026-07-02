# flow

One task at a time. For brains that need it.

`flow` is a CLI tool and AI server for people with ADHD, burnout, or anyone who finds conventional task managers overwhelming. It matches your tasks to your current energy level, never shows you a list when you just need one answer, and integrates with Claude, ChatGPT, Gemini, and any OpenAPI-aware AI so you can manage your day in plain conversation.

---

## Install

```bash
go install github.com/ajaykumarsingh/flow/cmd/flow@latest
```

Requires Go 1.21+. Single binary, no runtime dependencies, data stored locally at `~/.flow/flow.db`.

---

## Usage

### Get your next task

```bash
flow
```

Reads your last mood check-in and returns the single best task for your current energy. No list. One answer.

```
  → right now:

  reply to the thread [xs]
  id:  01KWHSKF…
  energy:  low

  flow done        — mark it complete
  flow focus        — start a focus timer
  flow break <id>   — break it into steps
```

### Add a task

```bash
flow add "write the quarterly review" --size l --energy high
flow add "reply to slack" --size xs --energy low
flow add "review the PR" --size s --energy med
```

`--size` — `xs s m l xl` (how much work it is)  
`--energy` — `low med high` (what state you need to be in)

### Check in your mood and energy

```bash
flow in
```

A quick interactive check-in (30 seconds). Your energy level gates which tasks get suggested — low energy surfaces only low-effort tasks, so you're never pushed into a mismatch.

### List your queue

```bash
flow ls         # pending tasks
flow ls --all   # include completed
```

### Mark done

```bash
flow done            # completes the currently suggested task
flow done 01KWHSKF   # complete a specific task by id prefix
```

### Break a task into steps

```bash
flow break 01KWHSKF "open the doc" "write the intro" "fill section 1" "review"
```

Creates subtasks linked to the parent. Each step is `xs` and inherits the parent's energy level.

### Capture a thought

```bash
flow note "too distracted to start, going for a walk"
```

Zero friction. No category, no tag required.

### Focus timer

```bash
flow focus               # 25 min on the suggested task
flow focus 01KWHSKF      # focus on a specific task
flow focus --minutes 45  # custom duration
```

Shows a live countdown. Ctrl+C ends the session early and records it as interrupted.

### Insights

```bash
flow insights
flow insights --days 14
```

Average mood and energy, task completion rate, and a 7-day mood chart.

---

## AI integrations

flow speaks two protocols — use whichever your AI supports.

---

## ChatGPT, Gemini, and any OpenAPI AI (`flow serve`)

`flow serve` starts a local REST API at `localhost:7777` and serves an OpenAPI 3.1 spec at `/openapi.json`. Any AI that supports OpenAPI tool calling can use it.

### Start the server

```bash
flow serve                          # no auth, port 7777
flow serve --api-key mysecretkey    # with Bearer token auth
FLOW_API_KEY=mysecretkey flow serve # via env var
flow serve --port 8080              # custom port
```

### Connect to ChatGPT (Custom GPT)

1. Run `flow serve --api-key <your-key>`
2. Expose it with a tunnel: `npx cloudflared tunnel --url http://localhost:7777`
3. In your Custom GPT → **Actions** → **Import from URL**: paste `https://<tunnel-url>/openapi.json`
4. Set auth: **API Key** → **Bearer** → your key
5. Ask it: *"I have 90 minutes and I'm exhausted — what should I work on?"*

### Connect to Gemini / Copilot / others

Any AI with OpenAPI support follows the same pattern — point it at `/openapi.json` and it discovers all available operations automatically.

### REST API endpoints

| Method | Path | What it does |
|---|---|---|
| `GET` | `/context` | Full state: tasks + last check-in + recent notes |
| `GET` | `/tasks` | List pending tasks |
| `POST` | `/tasks` | Add a task |
| `GET` | `/tasks/suggest` | Best task for current energy |
| `PUT` | `/tasks/{id}/complete` | Mark done |
| `POST` | `/tasks/{id}/breakdown` | Break into micro-steps |
| `POST` | `/checkins` | Log mood + energy |
| `POST` | `/notes` | Capture a thought |
| `GET` | `/insights` | Mood trends + completion stats |

---

## Claude integration (MCP)

`flow` runs as an MCP server, giving Claude full read/write access to your tasks, check-ins, and notes. Ask Claude to plan your day, break down tasks, or log how you're feeling — it all writes to the same local database your CLI uses.

### Setup

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on Mac):

```json
{
  "mcpServers": {
    "flow": {
      "command": "/Users/yourname/go/bin/flow",
      "args": ["mcp"]
    }
  }
}
```

Restart Claude Desktop. Then just talk to it:

> *"I have 90 minutes and I'm running at about 30% today — what should I work on?"*

> *"Add a task: prep for the 3pm call, medium size, medium energy"*

> *"I just finished the report. Mark it done and tell me what's next."*

> *"Break down 'write the proposal' into steps I can actually start"*

### MCP tools available to Claude

| Tool | What it does |
|---|---|
| `get_context` | Full current state: tasks, latest check-in, recent notes |
| `add_task` | Capture a task |
| `suggest_task` | Best task for current energy |
| `checkin` | Log mood + energy from conversation |
| `breakdown_task` | Split a task into micro-steps |
| `complete_task` | Mark a task done |
| `add_note` | Save a thought to the journal |
| `get_insights` | Mood trends and completion stats |

---

## How task suggestion works

`flow` never shows you a list and asks you to choose. The `flow` command (and `suggest_task` MCP tool) runs this logic:

1. Read your latest check-in (within the last 4 hours). Default to energy 3/5 if none.
2. Filter out tasks that require more energy than you have right now.
3. Score remaining tasks: prefer smaller tasks when energy is low, give a boost to anything already in progress, nudge older tasks slightly to prevent infinite deferral.
4. Return the top one.

If your queue is empty or nothing matches your energy, it tells you — and always suggests a next move.

---

## Data

Everything is stored in a single SQLite file at `~/.flow/flow.db`. No cloud, no account, no telemetry.

---

## License

MIT
