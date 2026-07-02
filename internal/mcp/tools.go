package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ajaykumarsingh/flow/internal/focus"
	"github.com/ajaykumarsingh/flow/internal/journal"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/task"
	mcpgo "github.com/mark3labs/mcp-go/server"
	"github.com/mark3labs/mcp-go/mcp"
)

func registerTools(srv *mcpgo.MCPServer, taskSvc *task.Service, moodSvc *mood.Service, journalSvc *journal.Service, focusSvc *focus.Service) {

	// ── get_context ──────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription("Returns the user's current state: pending tasks, latest check-in, recent notes. Call this first before giving any advice."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tasks, _ := taskSvc.List(false)
		checkin, _ := moodSvc.Latest(4 * time.Hour)
		notes, _ := journalSvc.Recent(5)

		type ctxResult struct {
			Tasks   []map[string]any `json:"tasks"`
			Checkin any              `json:"latest_checkin"`
			Notes   []map[string]any `json:"recent_notes"`
		}

		var taskList []map[string]any
		for _, t := range tasks {
			taskList = append(taskList, map[string]any{
				"id": t.ID, "title": t.Title,
				"size": t.Size, "energy": t.Energy, "status": t.Status,
			})
		}

		var checkinMap any
		if checkin != nil {
			checkinMap = map[string]any{
				"mood": checkin.Mood, "energy": checkin.Energy,
				"note": checkin.Note, "timestamp": checkin.Timestamp,
			}
		}

		var noteList []map[string]any
		for _, n := range notes {
			noteList = append(noteList, map[string]any{
				"id": n.ID, "content": n.Content, "timestamp": n.Timestamp,
			})
		}

		out, _ := json.MarshalIndent(ctxResult{Tasks: taskList, Checkin: checkinMap, Notes: noteList}, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	// ── add_task ─────────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("add_task",
		mcp.WithDescription("Add a new task. Infer size (xs/s/m/l/xl) and energy (low/med/high) from context if not provided."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("size", mcp.Description("xs s m l xl — defaults to m")),
		mcp.WithString("energy", mcp.Description("low med high — defaults to med")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.GetString("title", "")
		size := task.Size(req.GetString("size", "m"))
		energy := task.Energy(req.GetString("energy", "med"))
		tagsStr := req.GetString("tags", "")
		var tags []string
		if tagsStr != "" {
			for _, t := range strings.Split(tagsStr, ",") {
				if s := strings.TrimSpace(t); s != "" {
					tags = append(tags, s)
				}
			}
		}
		t, err := taskSvc.Add(title, size, energy, tags, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Added task %s: %s", t.ID[:8], t.Title)), nil
	})

	// ── suggest_task ─────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("suggest_task",
		mcp.WithDescription("Returns the single best task for the user's current energy level. Use when asked 'what should I work on?'"),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		checkin, _ := moodSvc.Latest(4 * time.Hour)
		energy := 3
		if checkin != nil {
			energy = checkin.Energy
		}
		tasks, _ := taskSvc.List(false)
		suggested := task.Suggest(tasks, energy)
		if suggested == nil {
			return mcp.NewToolResultText("No tasks in queue matching current energy."), nil
		}
		out, _ := json.Marshal(map[string]any{
			"id": suggested.ID, "title": suggested.Title,
			"size": suggested.Size, "energy": suggested.Energy,
		})
		return mcp.NewToolResultText(string(out)), nil
	})

	// ── checkin ───────────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("checkin",
		mcp.WithDescription("Log the user's mood and energy from conversation. Ask how they're feeling and record it."),
		mcp.WithNumber("mood", mcp.Required(), mcp.Description("Mood 1–5")),
		mcp.WithNumber("energy", mcp.Required(), mcp.Description("Energy 1–5")),
		mcp.WithString("note", mcp.Description("Optional note")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		moodVal := req.GetInt("mood", 3)
		energyVal := req.GetInt("energy", 3)
		note := req.GetString("note", "")
		c, err := moodSvc.Save(moodVal, energyVal, note)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Check-in logged: mood %d/5, energy %d/5 at %s", c.Mood, c.Energy, c.Timestamp.Format("15:04"))), nil
	})

	// ── breakdown_task ────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("breakdown_task",
		mcp.WithDescription("Break a task into ordered micro-steps. Generate the steps using your own reasoning, then store them."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("ID of the parent task")),
		mcp.WithString("steps", mcp.Required(), mcp.Description("JSON array of step titles, e.g. [\"open doc\",\"write intro\"]")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID := req.GetString("task_id", "")
		stepsJSON := req.GetString("steps", "[]")

		parent, err := taskSvc.Get(taskID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("task not found: %s", taskID)), nil
		}

		var steps []string
		if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
			return mcp.NewToolResultError("steps must be a JSON array of strings"), nil
		}

		var created []string
		for _, step := range steps {
			sub, err := taskSvc.Add(step, task.SizeXS, parent.Energy, nil, &parent.ID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			created = append(created, sub.Title)
		}
		return mcp.NewToolResultText(fmt.Sprintf("Created %d subtasks for '%s': %s", len(created), parent.Title, strings.Join(created, " · "))), nil
	})

	// ── complete_task ─────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("complete_task",
		mcp.WithDescription("Mark a task as done."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID := req.GetString("task_id", "")
		t, err := taskSvc.Get(taskID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("task not found: %s", taskID)), nil
		}
		if err := taskSvc.Complete(t.ID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("✓ Completed: %s", t.Title)), nil
	})

	// ── add_note ──────────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("add_note",
		mcp.WithDescription("Save a quick note or thought from the conversation into the journal."),
		mcp.WithString("content", mcp.Required(), mcp.Description("Note content")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content := req.GetString("content", "")
		tagsStr := req.GetString("tags", "")
		var tags []string
		if tagsStr != "" {
			for _, t := range strings.Split(tagsStr, ",") {
				if s := strings.TrimSpace(t); s != "" {
					tags = append(tags, s)
				}
			}
		}
		n, err := journalSvc.Add(content, tags)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Note saved: %s", n.ID[:8])), nil
	})

	// ── get_insights ──────────────────────────────────────────────────────────
	srv.AddTool(mcp.NewTool("get_insights",
		mcp.WithDescription("Returns mood trends, average energy, and task completion stats."),
		mcp.WithNumber("days", mcp.Description("Lookback period in days, defaults to 7")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		days := req.GetInt("days", 7)
		if days <= 0 {
			days = 7
		}
		checkins, _ := moodSvc.Recent(days * 3)
		tasks, _ := taskSvc.List(true)

		var moodSum, energySum int
		for _, c := range checkins {
			moodSum += c.Mood
			energySum += c.Energy
		}

		cutoff := time.Now().AddDate(0, 0, -days)
		var done, total int
		for _, t := range tasks {
			if t.CreatedAt.After(cutoff) {
				total++
				if t.Status == task.StatusDone {
					done++
				}
			}
		}

		result := map[string]any{
			"period_days":     days,
			"checkin_count":   len(checkins),
			"tasks_total":     total,
			"tasks_completed": done,
		}
		if len(checkins) > 0 {
			result["avg_mood"] = float64(moodSum) / float64(len(checkins))
			result["avg_energy"] = float64(energySum) / float64(len(checkins))
		}

		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}

