package mcp

import (
	"github.com/ajaykumarsingh/flow/internal/focus"
	"github.com/ajaykumarsingh/flow/internal/journal"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func NewServer(db *store.DB) *mcpgo.MCPServer {
	srv := mcpgo.NewMCPServer("flow", "1.0.0")

	taskSvc := task.NewService(db)
	moodSvc := mood.NewService(db)
	journalSvc := journal.NewService(db)
	focusSvc := focus.NewService(db)

	registerTools(srv, taskSvc, moodSvc, journalSvc, focusSvc)
	return srv
}
