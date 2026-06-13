package cli

import (
	"github.com/spf13/cobra"
)

// feedCmd returns the `feed <section>` command.
func (a *App) feedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "feed <section>",
		Short: "Fetch articles from a named InfoQ section feed",
		Long: `Fetch articles from a named InfoQ section feed.

Available sections: architecture-design, development, devops, ai-ml-data-eng,
software-vendors, culture-methods, java, dotnet, javascript.

Run "infoq sections" to list all sections with their URLs.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			section := args[0]
			n := a.effectiveLimit(20)
			a.progressf("fetching %s feed...", section)
			arts, err := a.client.Feed(cmd.Context(), section, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(arts, len(arts))
		},
	}
}
