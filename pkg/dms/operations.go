package dms

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

// ListTasks retrieves all DMS replication tasks
func (c *Client) ListTasks(ctx context.Context) ([]Task, error) {
	input := &databasemigrationservice.DescribeReplicationTasksInput{}

	var tasks []Task
	paginator := databasemigrationservice.NewDescribeReplicationTasksPaginator(c.svc, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", err)
		}

		for _, task := range output.ReplicationTasks {
			tasks = append(tasks, convertTask(task))
		}
	}

	return tasks, nil
}

// DescribeTask retrieves detailed information about a specific task
func (c *Client) DescribeTask(ctx context.Context, arn string) (*Task, error) {
	input := &databasemigrationservice.DescribeReplicationTasksInput{
		Filters: []types.Filter{
			{
				Name:   stringPtr("replication-task-arn"),
				Values: []string{arn},
			},
		},
	}

	output, err := c.svc.DescribeReplicationTasks(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe task: %w", err)
	}

	if len(output.ReplicationTasks) == 0 {
		return nil, fmt.Errorf("task not found: %s", arn)
	}

	task := convertTask(output.ReplicationTasks[0])
	return &task, nil
}

// GetTableStatistics retrieves table statistics for a specific task
func (c *Client) GetTableStatistics(ctx context.Context, arn string) ([]TableStatistic, error) {
	input := &databasemigrationservice.DescribeTableStatisticsInput{
		ReplicationTaskArn: stringPtr(arn),
	}

	var stats []TableStatistic
	paginator := databasemigrationservice.NewDescribeTableStatisticsPaginator(c.svc, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get table statistics: %w", err)
		}

		for _, stat := range output.TableStatistics {
			stats = append(stats, convertTableStatistic(stat))
		}
	}

	return stats, nil
}

// StartTask starts a DMS replication task
func (c *Client) StartTask(ctx context.Context, arn string, startType types.StartReplicationTaskTypeValue) error {
	input := &databasemigrationservice.StartReplicationTaskInput{
		ReplicationTaskArn:       stringPtr(arn),
		StartReplicationTaskType: startType,
	}

	_, err := c.svc.StartReplicationTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	return nil
}

// StopTask stops a DMS replication task
func (c *Client) StopTask(ctx context.Context, arn string) error {
	input := &databasemigrationservice.StopReplicationTaskInput{
		ReplicationTaskArn: stringPtr(arn),
	}

	_, err := c.svc.StopReplicationTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop task: %w", err)
	}

	return nil
}

// StartTasks starts multiple tasks in parallel
func (c *Client) StartTasks(ctx context.Context, arns []string, startType types.StartReplicationTaskTypeValue) []TaskOperation {
	var wg sync.WaitGroup
	results := make([]TaskOperation, len(arns))

	for i, arn := range arns {
		wg.Add(1)
		go func(index int, taskARN string) {
			defer wg.Done()

			err := c.StartTask(ctx, taskARN, startType)
			results[index] = TaskOperation{
				TaskARN: taskARN,
				Success: err == nil,
				Error:   err,
				Message: getOperationMessage("start", err),
			}
		}(i, arn)
	}

	wg.Wait()
	return results
}

// StopTasks stops multiple tasks in parallel
func (c *Client) StopTasks(ctx context.Context, arns []string) []TaskOperation {
	var wg sync.WaitGroup
	results := make([]TaskOperation, len(arns))

	for i, arn := range arns {
		wg.Add(1)
		go func(index int, taskARN string) {
			defer wg.Done()

			err := c.StopTask(ctx, taskARN)
			results[index] = TaskOperation{
				TaskARN: taskARN,
				Success: err == nil,
				Error:   err,
				Message: getOperationMessage("stop", err),
			}
		}(i, arn)
	}

	wg.Wait()
	return results
}

// RestartTask restarts a task (stop then start)
func (c *Client) RestartTask(ctx context.Context, arn string, startType types.StartReplicationTaskTypeValue) error {
	// First stop the task
	if err := c.StopTask(ctx, arn); err != nil {
		return fmt.Errorf("failed to stop task during restart: %w", err)
	}

	// Wait for task to stop (in production, you'd want to poll the status)
	// For now, we'll just start it immediately

	// Then start it
	if err := c.StartTask(ctx, arn, startType); err != nil {
		return fmt.Errorf("failed to start task during restart: %w", err)
	}

	return nil
}

// RestartTasks restarts multiple tasks in parallel
func (c *Client) RestartTasks(ctx context.Context, arns []string, startType types.StartReplicationTaskTypeValue) []TaskOperation {
	var wg sync.WaitGroup
	results := make([]TaskOperation, len(arns))

	for i, arn := range arns {
		wg.Add(1)
		go func(index int, taskARN string) {
			defer wg.Done()

			err := c.RestartTask(ctx, taskARN, startType)
			results[index] = TaskOperation{
				TaskARN: taskARN,
				Success: err == nil,
				Error:   err,
				Message: getOperationMessage("restart", err),
			}
		}(i, arn)
	}

	wg.Wait()
	return results
}

// Helper functions

func convertTask(task types.ReplicationTask) Task {
	t := Task{
		ARN:                    stringValue(task.ReplicationTaskArn),
		Name:                   stringValue(task.ReplicationTaskIdentifier),
		Status:                 stringValue(task.Status),
		ReplicationInstanceARN: stringValue(task.ReplicationInstanceArn),
		SourceEndpointARN:      stringValue(task.SourceEndpointArn),
		TargetEndpointARN:      stringValue(task.TargetEndpointArn),
		MigrationType:          string(task.MigrationType),
		TableMappings:          stringValue(task.TableMappings),
		CreatedAt:              task.ReplicationTaskCreationDate,
		StartedAt:              task.ReplicationTaskStartDate,
		StoppedAt:              nil, // AWS SDK doesn't provide a stopped timestamp
		LastFailureMessage:     stringValue(task.LastFailureMessage),
	}

	if task.ReplicationTaskStats != nil {
		t.ReplicationTaskStats = &TaskStats{
			FullLoadProgressPercent: task.ReplicationTaskStats.FullLoadProgressPercent,
			ElapsedTimeMillis:       task.ReplicationTaskStats.ElapsedTimeMillis,
			TablesLoaded:            task.ReplicationTaskStats.TablesLoaded,
			TablesLoading:           task.ReplicationTaskStats.TablesLoading,
			TablesQueued:            task.ReplicationTaskStats.TablesQueued,
			TablesErrored:           task.ReplicationTaskStats.TablesErrored,
			StopReason:              stringValue(task.StopReason),
		}
	}

	return t
}

func convertTableStatistic(stat types.TableStatistics) TableStatistic {
	return TableStatistic{
		SchemaName:      stringValue(stat.SchemaName),
		TableName:       stringValue(stat.TableName),
		Inserts:         stat.Inserts,
		Deletes:         stat.Deletes,
		Updates:         stat.Updates,
		Ddls:            stat.Ddls,
		FullLoadRows:    stat.FullLoadRows,
		LastUtctime:     stat.LastUpdateTime,
		ValidationState: stringValue(stat.ValidationState),
	}
}

func getOperationMessage(operation string, err error) string {
	if err != nil {
		return fmt.Sprintf("Failed to %s task: %v", operation, err)
	}
	return fmt.Sprintf("Successfully issued %s command", operation)
}

func stringPtr(s string) *string {
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
