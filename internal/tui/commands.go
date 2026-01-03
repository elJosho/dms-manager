package tui

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eljosho/dms-manager/pkg/dms"
)

// Messages for async operations

type tasksLoadedMsg struct {
	tasks []dms.Task
	err   error
}

type taskOperationCompleteMsg struct {
	results []dms.TaskOperation
}

type errorMsg struct {
	err error
}

const refreshInterval = 5 * time.Second

// LoadTasksCmd loads tasks asynchronously
func LoadTasksCmd(client *dms.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		tasks, err := client.ListTasks(ctx)
		return tasksLoadedMsg{tasks: tasks, err: err}
	}
}

// StartTasksCmd starts tasks asynchronously
func StartTasksCmd(client *dms.Client, arns []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// Default to start-replication type
		results := client.StartTasks(ctx, arns, types.StartReplicationTaskTypeValueStartReplication)
		return taskOperationCompleteMsg{results: results}
	}
}

// StopTasksCmd stops tasks asynchronously
func StopTasksCmd(client *dms.Client, arns []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := client.StopTasks(ctx, arns)
		return taskOperationCompleteMsg{results: results}
	}
}

// ResumeTasksCmd resumes tasks asynchronously
func ResumeTasksCmd(client *dms.Client, arns []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := client.StartTasks(ctx, arns, types.StartReplicationTaskTypeValueResumeProcessing)
		return taskOperationCompleteMsg{results: results}
	}
}

// ReloadTasksCmd reloads tasks asynchronously (stop then start with reload-target)
func ReloadTasksCmd(client *dms.Client, arns []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := client.RestartTasks(ctx, arns, types.StartReplicationTaskTypeValueReloadTarget)
		return taskOperationCompleteMsg{results: results}
	}
}

type tableStatsLoadedMsg struct {
	stats []dms.TableStatistic
	err   error
}

// LoadTableStatsCmd loads table statistics asynchronously
func LoadTableStatsCmd(client *dms.Client, arn string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		stats, err := client.GetTableStatistics(ctx, arn)
		return tableStatsLoadedMsg{stats: stats, err: err}
	}
}

// TickCmd returns a command that waits and then sends a tick message
func TickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}
