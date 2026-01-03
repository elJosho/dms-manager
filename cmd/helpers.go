package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/eljosho/dms-manager/pkg/dms"
)

// resolveTaskARNs converts a mix of ARNs and task names to ARNs
func resolveTaskARNs(ctx context.Context, client *dms.Client, identifiers []string) ([]string, error) {
	taskARNs := make([]string, 0, len(identifiers))

	for _, id := range identifiers {
		if strings.HasPrefix(id, "arn:") {
			// Already an ARN
			taskARNs = append(taskARNs, id)
		} else {
			// Need to find task by name
			tasks, err := client.ListTasks(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list tasks: %w", err)
			}

			found := false
			for _, task := range tasks {
				if task.Name == id {
					taskARNs = append(taskARNs, task.ARN)
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("Warning: task '%s' not found\n", id)
			}
		}
	}

	return taskARNs, nil
}

// getTaskNameFromARN extracts a readable name from an ARN
func getTaskNameFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) > 0 {
		// Get the last part which usually contains the task name
		lastPart := parts[len(parts)-1]
		// Remove any prefix like "task:"
		if strings.Contains(lastPart, "/") {
			subparts := strings.Split(lastPart, "/")
			return subparts[len(subparts)-1]
		}
		return lastPart
	}
	return arn
}
