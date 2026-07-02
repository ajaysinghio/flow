# flow

One task at a time. For brains that need it.

`flow` is a CLI tool and AI server for people with ADHD, burnout, or anyone who finds conventional task managers overwhelming. It matches your tasks to your current energy level, never shows you a list when you just need one answer, and integrates with Claude, ChatGPT, Gemini, and any OpenAPI-aware AI so you can manage your day in plain conversation.

It runs in four ways — pick whatever has the least friction right now:

| Interface | Command | Best for |
|---|---|---|
| Terminal | `flow` | Developers, quick captures |
| Menu bar | `flow tray` | Always-on, glanceable |
| Claude (MCP) | `flow mcp` | Conversational planning |
| ChatGPT / Gemini | `flow serve` | OpenAPI tool calling |

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

Reads your last check-in and returns the single best task for your current energy. No list. One answer.

```
  → right now:

  reply to the thread [xs]
  id:  01KWHSKF…
  energy:  low

  flow done        — mark it complete
  flow focus        — start a focus timer
  flow break <id>   — break it into steps
```

If your energy is too low for any task in your queue, flow tells you why and suggests next steps rather than showing an empty screen:

```
  No tasks match your current energy (1/5).
  Try: flow pick  to choose from your full list
       flow in   to update your energy level
```

---

### Add a task

```bash
flow add "call the dentist"
flow add "finish the report" --due friday
flow add "renew subscription" --due 2026-07-10
flow add "prep for standup" --due tomorrow
```

No required flags. Just write the task. `--due` is the only optional flag and accepts natural language:

| Input | Means |
|---|---|
| `today` | end of today |
| `tomorrow` | end of tomorrow |
| `friday` | end of next Friday |
| `next week` | 7 days from now |
| `2026-07-10` | specific date |

---

### Pick a task, focus, done

```bash
flow pick
flow pick --minutes 45
```

The full loop in one command. Shows your queue ranked by urgency and energy, you pick one, a focus timer starts, and when time's up you're asked whether to mark it done.

```
  Pick a task  ↑↓ navigate  Enter to start  q to cancel

  ▶ reply to the thread          ← cursor
    finish the report  (friday)
    call the dentist   (today)
    prep for standup
```

This is the recommended way to start a work block.

---

### Check in your mood and energy

```bash
flow in
```

A quick interactive check-in (30 seconds). Your energy level gates which tasks get suggested — low energy surfaces only low-effort tasks, so you're never pushed into a mismatch.

---

### List your queue

```bash
flow ls         # pending tasks
flow ls --all   # include completed
```

Due dates are shown inline with urgency indicators:

```
  ○  finish the report   ⚠ overdue    01KWHSKF…
  ○  call the dentist    ⏰ today      01KWHSKF…
  ○  review the PR       fri          01KWHSKF…
  ○  reply to thread                  01KWHSKF…
```

---

### Mark done

```bash
flow done            # completes the currently suggested task
flow done 01KWHSKF   # complete a specific task by id prefix
```

---

### Break a task into steps

```bash
flow break 01KWHSKF "open the doc" "write the intro" "fill section 1" "review"
```

Creates subtasks linked to the parent. Each step is `xs` sized and inherits the parent's energy level.

---

### Focus timer

```bash
flow focus               # 25 min on the suggested task
flow focus 01KWHSKF      # focus on a specific task
flow focus --minutes 45  # custom duration
```

Shows a live countdown. Ctrl+C ends the session early and records it as interrupted. Prefer `flow pick` if you haven't chosen a task yet — it combines picking, focusing, and marking done in one flow.

---

### Capture a thought

```bash
flow note "too distracted to start, going for a walk"
```

Zero friction. No category, no tag required.

---

### Insights

```bash
flow insights
flow insights --days 14
```

Average mood and energy, task completion rate, and a 7-day mood chart.

---

## macOS menu bar / Windows system tray

```bash
flow tray
```

Adds flow to your macOS menu bar (or Windows system tray). Shows your current suggested task and lets you act on it without opening a terminal.

```
  →  reply to the thread
  ✓  Mark done
  ↻  Refresh
  ──────────────────────
  ≡  Pick a task  ▶  reply to the thread
                     finish the report  (friday)
                     call the dentist   (today)
                     review the PR
                     prep for standup
  +  Add task…
  ◎  Check in    ▶  1  drained
                    2  low
                    3  medium
                    4  good
                    5  charged
  ──────────────────────
  Quit flow
```

- Suggested task updates automatically every 30 seconds
- **Pick a task** submenu shows the top 5 ranked tasks with due dates — click any to make it current
- **Add task** opens a native macOS dialog box — no terminal needed
- **Check in** sets your energy instantly, which immediately affects what gets suggested
- All actions write to the same `~/.flow/flow.db` as the CLI and AI integrations

To run on login, add it to macOS Login Items (`System Settings → General → Login Items`).

---

## AI integrations

flow speaks two protocols — use whichever your AI supports.

---

### ChatGPT, Gemini, and any OpenAPI AI (`flow serve`)

`flow serve` starts a local REST API at `localhost:7777` and serves an OpenAPI 3.1 spec at `/openapi.json`. Any AI that supports OpenAPI tool calling can use it.

```bash
flow serve                          # no auth, port 7777
flow serve --api-key mysecretkey    # with Bearer token auth
FLOW_API_KEY=mysecretkey flow serve # via env var
flow serve --port 8080              # custom port
```

**Connect to ChatGPT (Custom GPT)**

1. Run `flow serve --api-key <your-key>`
2. Expose it with a tunnel: `npx cloudflared tunnel --url http://localhost:7777`
3. In your Custom GPT → **Actions** → **Import from URL**: paste `https://<tunnel-url>/openapi.json`
4. Set auth: **API Key** → **Bearer** → your key
5. Ask it: *"I have 90 minutes and I'm exhausted — what should I work on?"*

**Connect to Gemini / Copilot / others**

Any AI with OpenAPI support follows the same pattern — point it at `/openapi.json` and it discovers all available operations automatically.

**REST API endpoints**

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

### Claude integration (MCP)

`flow` runs as an MCP server, giving Claude full read/write access to your tasks, check-ins, and notes.

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

> *"Add a task: prep for the 3pm call, due today"*

> *"I just finished the report. Mark it done and tell me what's next."*

> *"Break down 'write the proposal' into steps I can actually start"*

**MCP tools available to Claude**

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

`flow` never asks you to choose. The `flow` command runs this logic:

1. Read your latest check-in (within the last 4 hours). Default to energy 3/5 if none.
2. Filter out tasks that require more energy than you have right now.
3. Score remaining tasks:
   - **Overdue** tasks: +8 points — surfaces them first
   - **Due today**: +5 points
   - **Due this week**: +2 points
   - **In progress**: +4 points — momentum matters
   - **Smaller tasks** score higher when energy is low
   - **Older tasks** get a mild nudge to prevent infinite deferral
4. Return the top one.

If your energy is too low for any task in your queue, flow says so and tells you what to do — it never shows a blank screen.

---

## Data

Everything is stored in a single SQLite file at `~/.flow/flow.db`. No cloud, no account, no telemetry.

---

## License

MIT
