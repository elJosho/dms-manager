package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [task-arn-or-name...]",
	Short: "Resume one or more DMS replication tasks",
	Long: `Resume one or more DMS replication tasks that were previously stopped.
	
You can specify multiple task ARNs or task names as arguments.
Tasks will be resumed concurrently for faster execution.

Wildcards are supported for task names (e.g. "prod-*", "*-database").
Note: When using wildcards, you MUST quote the argument to prevent shell expansion.

Examples:
  dms-manager resume task1 task2
  dms-manager resume "*-database"

This uses the resume-processing start type, which resumes replication 
from where it was stopped.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runResume,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) {
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

	fmt.Printf("Resuming %d task(s) in parallel...\n\n", len(taskARNs))

	results := client.StartTasks(ctx, taskARNs, types.StartReplicationTaskTypeValueResumeProcessing)

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

	fmt.Printf("\nSuccessfully resumed %d out of %d tasks\n", successCount, len(taskARNs))
}
