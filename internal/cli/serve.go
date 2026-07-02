package cli

import (
	"fmt"
	"os"

	"github.com/ajaykumarsingh/flow/internal/api"
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/spf13/cobra"
)

func newServeCmd(a *app.App) *cobra.Command {
	var port int
	var apiKey string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the REST API server (for ChatGPT, Gemini, and any OpenAPI-aware AI)",
		Long: `Starts flow as a local REST API server with an OpenAPI 3.1 spec.

The OpenAPI spec is served at /openapi.json — paste that URL into your
Custom GPT Actions config or any OpenAPI-compatible AI tool.

To expose locally to external AI services, use a tunnel:
  npx cloudflared tunnel --url http://localhost:7777
  ngrok http 7777`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := apiKey
			if key == "" {
				key = os.Getenv("FLOW_API_KEY")
			}
			addr := fmt.Sprintf("localhost:%d", port)
			srv := api.NewServer(a, key)

			fmt.Printf("\n  %s\n", styleAccent.Render("flow API server"))
			fmt.Printf("  %s  http://%s\n", styleDim.Render("listening:"), addr)
			fmt.Printf("  %s  http://%s/openapi.json\n", styleDim.Render("spec:"), addr)
			if key != "" {
				fmt.Printf("  %s  Bearer token required\n", styleDim.Render("auth:"))
			} else {
				fmt.Printf("  %s  %s\n", styleDim.Render("auth:"), styleAccent.Render("none — set --api-key or FLOW_API_KEY"))
			}
			fmt.Println()
			return srv.Listen(addr)
		},
	}

	cmd.Flags().IntVar(&port, "port", 7777, "Port to listen on")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Bearer token (or set FLOW_API_KEY env var)")
	return cmd
}
