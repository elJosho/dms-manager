package cmd

import (
	"context"
	"fmt"

	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [task-arn-or-name...]",
	Short: "Stop one or more DMS replication tasks",
	Long: `Stop one or more DMS replication tasks in parallel.
	
You can specify multiple task ARNs or task names as arguments.
Tasks will be stopped concurrently for faster execution.

Wildcards are supported for task names (e.g. "prod-*", "*-database").
Note: When using wildcards, you MUST quote the argument to prevent shell expansion.

Examples:
  dms-manager stop task1 task2
  dms-manager stop "*-database"
  dms-manager stop "prod-*"`,
	Args: cobra.MinimumNArgs(1),
	Run:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	client, err := dms.NewClient(ctx, GetProfile(), GetRegion())
	if err != nil {
		exitWithError(fmt.Errorf("failed to create DMS client: %w", err))
	}

	// Resolve task names to ARNs
	taskARNs, err := resolveTaskARNs(ctx, client, args)
	if err != nil {
		exitWithError(err)
	}

	if len(taskARNs) == 0 {
		exitWithError(fmt.Errorf("no valid tasks found"))
	}

	fmt.Printf("Stopping %d task(s) in parallel...\n\n", len(taskARNs))

	results := client.StopTasks(ctx, taskARNs)

	// Print results
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			fmt.Printf("✓ %s: %s\n", getTaskNameFromARN(result.TaskARN), result.Message)
		} else {
			fmt.Printf("✗ %s: %s\n", getTaskNameFromARN(result.TaskARN), result.Message)
		}
	}

	fmt.Printf("\nSuccessfully stopped %d out of %d tasks\n", successCount, len(taskARNs))
}
