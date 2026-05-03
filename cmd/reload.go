package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload [task-arn-or-name...]",
	Short: "Reload one or more DMS replication tasks",
	Long: `Reload one or more DMS replication tasks (stop then start with reload-target) in parallel.
	
You can specify multiple task ARNs or task names as arguments.
Tasks will be reloaded concurrently for faster execution.

Wildcards are supported for task names (e.g. "prod-*", "*-database").
Note: When using wildcards, you MUST quote the argument to prevent shell expansion.

Examples:
  dms-manager reload task1 task2
  dms-manager reload "*-database"
  dms-manager reload task1 --table schema.table`,
	Args: cobra.MinimumNArgs(1),
	Run:  runReload,
}

func init() {
	reloadCmd.Flags().String("table", "", "Specific table to reload in format 'schema.table'")
	rootCmd.AddCommand(reloadCmd)
}

func runReload(cmd *cobra.Command, args []string) {
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

	tableFlag, _ := cmd.Flags().GetString("table")

	var results []dms.TaskOperation
	if tableFlag != "" {
		parts := strings.Split(tableFlag, ".")
		if len(parts) != 2 {
			exitWithError(fmt.Errorf("invalid table format '%s', expected 'schema.table'", tableFlag))
		}
		fmt.Printf("Reloading table %s.%s on %d task(s) in parallel...\n\n", parts[0], parts[1], len(taskARNs))
		results = client.ReloadTablesInTasks(ctx, taskARNs, parts[0], parts[1])
	} else {
		fmt.Printf("Reloading %d task(s) in parallel...\n\n", len(taskARNs))
		// Use reload-target start type
		results = client.RestartTasks(ctx, taskARNs, types.StartReplicationTaskTypeValueReloadTarget)
	}

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

	fmt.Printf("\nSuccessfully reloaded %d out of %d tasks\n", successCount, len(taskARNs))
}
