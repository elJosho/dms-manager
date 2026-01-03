package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/eljosho/dms-manager/pkg/dms"
)

// expandWildcards checks if any identifier contains wildcards and expands to matching tasks
func expandWildcards(ctx context.Context, client *dms.Client, identifiers []string) ([]string, error) {
	// Check if any identifier contains wildcards or is "all"
	hasPattern := false
	for _, id := range identifiers {
		if id == "all" || strings.ContainsAny(id, "*?[") {
			hasPattern = true
			break
		}
	}

	// If no patterns, return original identifiers
	if !hasPattern {
		return identifiers, nil
	}

	// Fetch all tasks once
	tasks, err := client.ListTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Expand each identifier
	expanded := []string{}
	for _, id := range identifiers {
		if id == "all" || id == "*" {
			// Special case: match all tasks
			for _, task := range tasks {
				expanded = append(expanded, task.ARN)
			}
		} else if strings.ContainsAny(id, "*?[") {
			// Pattern matching using filepath.Match
			matchCount := 0
			for _, task := range tasks {
				// Use a simple glob pattern matcher
				matched := matchPattern(id, task.Name)

				if matched {
					expanded = append(expanded, task.ARN)
					matchCount++
				}
			}
			if matchCount == 0 {
				fmt.Printf("Warning: no tasks matched pattern '%s'\n", id)
			}
		} else {
			// No wildcards, keep as-is for name resolution
			expanded = append(expanded, id)
		}
	}

	return expanded, nil
}

// matchPattern performs simple glob pattern matching
func matchPattern(pattern, name string) bool {
	// Convert glob pattern to simple matching logic
	// * matches any sequence of characters
	// ? matches any single character

	// Simple implementation without filepath.Match to avoid path separator issues
	i, j := 0, 0
	starIdx, matchIdx := -1, 0

	for j < len(name) {
		if i < len(pattern) && (pattern[i] == name[j] || pattern[i] == '?') {
			i++
			j++
		} else if i < len(pattern) && pattern[i] == '*' {
			starIdx = i
			matchIdx = j
			i++
		} else if starIdx != -1 {
			i = starIdx + 1
			matchIdx++
			j = matchIdx
		} else {
			return false
		}
	}

	for i < len(pattern) && pattern[i] == '*' {
		i++
	}

	return i == len(pattern)
}

// resolveTaskARNs converts a mix of ARNs and task names to ARNs
func resolveTaskARNs(ctx context.Context, client *dms.Client, identifiers []string) ([]string, error) {
	// First, expand any wildcards
	expanded, err := expandWildcards(ctx, client, identifiers)
	if err != nil {
		return nil, err
	}

	taskARNs := make([]string, 0, len(expanded))

	for _, id := range expanded {
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
