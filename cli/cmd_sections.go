package cli

import (
	"github.com/spf13/cobra"
)

// sectionsCmd returns the `sections` command.
func (a *App) sectionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sections",
		Short: "List available InfoQ topic sections",
		RunE: func(cmd *cobra.Command, _ []string) error {
			secs := a.client.Sections()
			return a.renderOrEmpty(secs, len(secs))
		},
	}
}
