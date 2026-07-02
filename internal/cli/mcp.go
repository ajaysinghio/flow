package cli

import (
	"fmt"

	"github.com/ajaykumarsingh/flow/internal/app"
	flowmcp "github.com/ajaykumarsingh/flow/internal/mcp"
	mcpgo "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func newMCPCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server (stdio) for Claude integration",
		Long: `Starts flow as an MCP server on stdio.

Add to your Claude Desktop config:

  "mcpServers": {
    "flow": { "command": "flow", "args": ["mcp"] }
  }

Then ask Claude: "flow, I have 90 minutes at 40% energy — what should I work on?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := flowmcp.NewServer(a)
			fmt.Fprintln(cmd.ErrOrStderr(), "flow MCP server starting on stdio…")
			return mcpgo.NewStdioServer(srv).Listen(cmd.Context(), nil, nil)
		},
	}
}
