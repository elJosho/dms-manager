package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/eljosho/dms-manager/internal/tui"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var tablesCmd = &cobra.Command{
	Use:   "tables [task-arn-or-name...]",
	Short: "List out tables in one or more DMS replication tasks",
	Long: `List out all tables and their statistics for one or more DMS replication tasks.
	
You can specify multiple task ARNs or task names as arguments.
Wildcards are supported (e.g. "prod-*").`,
	Args: cobra.MinimumNArgs(1),
	Run:  runTables,
}

func init() {
	rootCmd.AddCommand(tablesCmd)
}

func runTables(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	client, err := dms.NewClient(ctx, GetProfile(), GetRegion())
	if err != nil {
		exitWithError(fmt.Errorf("failed to create DMS client: %w", err))
	}

	taskARNs, err := resolveTaskARNs(ctx, client, args)
	if err != nil {
		exitWithError(err)
	}

	if len(taskARNs) == 0 {
		exitWithError(fmt.Errorf("no valid tasks found"))
	}

	for i, arn := range taskARNs {
		if i > 0 {
			fmt.Println("\n" + tui.CLIMutedStyle.Render(strings.Repeat("─", 80)))
		}

		taskName := getTaskNameFromARN(arn)
		fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Task:"), tui.CLIPrimaryStyle.Render(taskName))

		stats, err := client.GetTableStatistics(ctx, arn)
		if err != nil {
			fmt.Printf("\n%s %v\n", tui.CLIErrorStyle.Render("Error fetching table statistics:"), err)
		} else {
			printTableStatistics(stats)
		}
	}
}
