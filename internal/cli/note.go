package cli

import (
	"fmt"
	"strings"

	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/spf13/cobra"
)

func newNoteCmd(a *app.App) *cobra.Command {
	var tags []string
	cmd := &cobra.Command{
		Use:     "note <text>",
		Short:   "Capture a thought instantly",
		Example: `  flow note "feeling foggy, can't start the report"`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := a.Journal.Add(strings.Join(args, " "), tags)
			if err != nil {
				return err
			}
			fmt.Printf("\n  %s  %s\n\n", styleGreen.Render("✓ noted"), styleDim.Render(n.ID[:8]+"…"))
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags")
	return cmd
}
