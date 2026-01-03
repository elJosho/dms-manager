package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var (
	startType string
)

var startCmd = &cobra.Command{
	Use:   "start [task-arn-or-name...]",
	Short: "Start one or more DMS replication tasks",
	Long: `Start one or more DMS replication tasks in parallel.
	
You can specify multiple task ARNs or task names as arguments.
Tasks will be started concurrently for faster execution.

Wildcards are supported for task names (e.g. "prod-*", "*-database").
Note: When using wildcards, you MUST quote the argument to prevent shell expansion.

Examples:
  dms-manager start task1 task2
  dms-manager start "*-database"
  dms-manager start "prod-*" --type resume-processing`,
	Args: cobra.MinimumNArgs(1),
	Run:  runStart,
}

func init() {
	startCmd.Flags().StringVarP(&startType, "type", "t", "start-replication", "Start type: start-replication, resume-processing, or reload-target")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) {
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

	// Parse start type
	var taskStartType types.StartReplicationTaskTypeValue
	switch strings.ToLower(startType) {
	case "start-replication":
		taskStartType = types.StartReplicationTaskTypeValueStartReplication
	case "resume-processing":
		taskStartType = types.StartReplicationTaskTypeValueResumeProcessing
	case "reload-target":
		taskStartType = types.StartReplicationTaskTypeValueReloadTarget
	default:
		exitWithError(fmt.Errorf("invalid start type: %s (use start-replication, resume-processing, or reload-target)", startType))
	}

	fmt.Printf("Starting %d task(s) in parallel...\n\n", len(taskARNs))

	results := client.StartTasks(ctx, taskARNs, taskStartType)

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

	fmt.Printf("\nSuccessfully started %d out of %d tasks\n", successCount, len(taskARNs))
}
