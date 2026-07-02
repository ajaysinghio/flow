// Package app is the core aggregate — every presentation layer (CLI, API, MCP, tray)
// takes *app.App instead of touching the store or individual services directly.
package app

import (
	"github.com/ajaykumarsingh/flow/internal/focus"
	"github.com/ajaykumarsingh/flow/internal/journal"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
)

type App struct {
	Tasks   *task.Service
	Moods   *mood.Service
	Journal *journal.Service
	Focus   *focus.Service
	DB      *store.DB
}

func New(db *store.DB) *App {
	return &App{
		DB:      db,
		Tasks:   task.NewService(db),
		Moods:   mood.NewService(db),
		Journal: journal.NewService(db),
		Focus:   focus.NewService(db),
	}
}
