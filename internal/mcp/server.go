package mcp

import (
	"github.com/ajaykumarsingh/flow/internal/app"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func NewServer(a *app.App) *mcpgo.MCPServer {
	srv := mcpgo.NewMCPServer("flow", "1.0.0")
	registerTools(srv, a.Tasks, a.Moods, a.Journal, a.Focus)
	return srv
}
