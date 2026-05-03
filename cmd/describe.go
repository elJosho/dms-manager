package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljosho/dms-manager/internal/tui"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe [task-arn...]",
	Short: "Get detailed information about DMS tasks",
	Long: `Get detailed information about one or more DMS replication tasks.
	
You can specify multiple task ARNs or task names as arguments.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDescribe,
}

func init() {
	describeCmd.Flags().Bool("tables", false, "Show table statistics")
	rootCmd.AddCommand(describeCmd)
}

func runDescribe(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	client, err := dms.NewClient(ctx, GetProfile(), GetRegion())
	if err != nil {
		exitWithError(fmt.Errorf("failed to create DMS client: %w", err))
	}

	// Resolve task names and wildcards to ARNs
	taskARNs, err := resolveTaskARNs(ctx, client, args)
	if err != nil {
		exitWithError(err)
	}

	if len(taskARNs) == 0 {
		exitWithError(fmt.Errorf("no valid tasks found"))
	}

	// Describe each task
	showTables, _ := cmd.Flags().GetBool("tables")

	for i, arn := range taskARNs {
		if i > 0 {
			fmt.Println("\n" + tui.CLIMutedStyle.Render(strings.Repeat("─", 80)))
		}

		task, err := client.DescribeTask(ctx, arn)
		if err != nil {
			fmt.Printf("%s %s: %v\n", tui.CLIErrorStyle.Render("Error describing task"), arn, err)
			continue
		}

		printTaskDetails(task)

		if showTables {
			stats, err := client.GetTableStatistics(ctx, arn)
			if err != nil {
				fmt.Printf("\n%s %v\n", tui.CLIErrorStyle.Render("Error fetching table statistics:"), err)
			} else {
				printTableStatistics(stats)
			}
		}
	}
}

func printTaskDetails(task *dms.Task) {
	// Task header
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Task:"), tui.CLIPrimaryStyle.Render(task.Name))
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("ARN:"), tui.CLIMutedStyle.Render(dms.TruncateARN(task.ARN, 60)))
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Status:"), getDescribeStatusStyle(task.Status).Render(task.Status))
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Migration Type:"), tui.CLIValueStyle.Render(task.MigrationType))

	// Endpoints section
	fmt.Println("\n" + tui.CLIHighlightStyle.Render("Endpoints:"))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Replication Instance:"), tui.CLIMutedStyle.Render(dms.TruncateARN(task.ReplicationInstanceARN, 60)))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Source:"), tui.CLIMutedStyle.Render(dms.TruncateARN(task.SourceEndpointARN, 60)))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Target:"), tui.CLIMutedStyle.Render(dms.TruncateARN(task.TargetEndpointARN, 60)))

	// Timestamps
	if task.CreatedAt != nil {
		fmt.Printf("\n%s %s\n", tui.CLILabelStyle.Render("Created At:"), tui.CLIValueStyle.Render(task.CreatedAt.Format("2006-01-02 15:04:05")))
	}

	if task.StartedAt != nil {
		fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Started At:"), tui.CLIValueStyle.Render(task.StartedAt.Format("2006-01-02 15:04:05")))
	}

	if task.LastFailureMessage != "" {
		fmt.Printf("\n%s %s\n", tui.CLIErrorStyle.Render("Last Failure:"), task.LastFailureMessage)
	}

	if task.ReplicationTaskStats != nil {
		stats := task.ReplicationTaskStats
		fmt.Println("\n" + tui.CLIHighlightStyle.Render("Statistics:"))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Full Load Progress:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d%%", stats.FullLoadProgressPercent)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loaded:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesLoaded)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loading:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesLoading)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Queued:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesQueued)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Errored:"), getDescribeErrorCountStyle(stats.TablesErrored).Render(dms.FormatNumber(stats.TablesErrored)))

		if stats.ElapsedTimeMillis > 0 {
			fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Elapsed Time:"), tui.CLIValueStyle.Render(dms.FormatElapsedTime(stats.ElapsedTimeMillis)))
		}

		if stats.StopReason != "" {
			fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Stop Reason:"), tui.CLIWarningStyle.Render(stats.StopReason))
		}
	}
}

func printTableStatistics(stats []dms.TableStatistic) {
	if len(stats) == 0 {
		fmt.Println("\n" + tui.CLIWarningStyle.Render("Table Statistics: None"))
		return
	}

	fmt.Println("\n" + tui.CLIHighlightStyle.Render("Table Statistics:"))
	// Header
	headerFmt := "  %-15s %-30s %-10s %-10s %-10s %-8s %-10s %-15s"
	fmt.Printf(headerFmt+"\n",
		tui.CLIHeaderStyle.Render("SCHEMA"),
		tui.CLIHeaderStyle.Render("TABLE"),
		tui.CLIHeaderStyle.Render("INSERTS"),
		tui.CLIHeaderStyle.Render("UPDATES"),
		tui.CLIHeaderStyle.Render("DELETES"),
		tui.CLIHeaderStyle.Render("DDLS"),
		tui.CLIHeaderStyle.Render("ROWS"),
		tui.CLIHeaderStyle.Render("STATE"),
	)

	// Separator
	fmt.Printf(headerFmt+"\n",
		tui.CLIMutedStyle.Render("───────────────"),
		tui.CLIMutedStyle.Render("──────────────────────────────"),
		tui.CLIMutedStyle.Render("──────────"),
		tui.CLIMutedStyle.Render("──────────"),
		tui.CLIMutedStyle.Render("──────────"),
		tui.CLIMutedStyle.Render("────────"),
		tui.CLIMutedStyle.Render("──────────"),
		tui.CLIMutedStyle.Render("───────────────"),
	)

	for _, s := range stats {
		stateStyle := getValidationStateStyle(s.ValidationState)

		schemaPad := fmt.Sprintf("%-15s", truncateString(s.SchemaName, 15))
		tablePad := fmt.Sprintf("%-30s", truncateString(s.TableName, 30))
		insertsPad := fmt.Sprintf("%-10s", dms.FormatNumber(s.Inserts))
		updatesPad := fmt.Sprintf("%-10s", dms.FormatNumber(s.Updates))
		deletesPad := fmt.Sprintf("%-10s", dms.FormatNumber(s.Deletes))
		ddlsPad := fmt.Sprintf("%-8s", dms.FormatNumber(s.Ddls))
		rowsPad := fmt.Sprintf("%-10s", dms.FormatNumber(s.FullLoadRows))
		statePad := fmt.Sprintf("%-15s", s.ValidationState)

		fmt.Printf("  %s %s %s %s %s %s %s %s\n",
			tui.CLIValueStyle.Render(schemaPad),
			tui.CLIPrimaryStyle.Render(tablePad),
			tui.CLINumberStyle.Render(insertsPad),
			tui.CLINumberStyle.Render(updatesPad),
			tui.CLINumberStyle.Render(deletesPad),
			tui.CLINumberStyle.Render(ddlsPad),
			tui.CLINumberStyle.Render(rowsPad),
			stateStyle.Render(statePad),
		)
	}
}

// getDescribeStatusStyle returns the appropriate style for a task status
func getDescribeStatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "running", "starting", "replicating":
		return tui.CLISuccessStyle
	case "stopped", "stopping", "failed":
		return tui.CLIErrorStyle
	default:
		return tui.CLIWarningStyle
	}
}

// getDescribeErrorCountStyle returns red for non-zero error counts
func getDescribeErrorCountStyle(count int32) lipgloss.Style {
	if count > 0 {
		return tui.CLIErrorStyle
	}
	return tui.CLINumberStyle
}

// getValidationStateStyle returns color based on validation state
func getValidationStateStyle(state string) lipgloss.Style {
	switch strings.ToLower(state) {
	case "validated", "table validated":
		return tui.CLISuccessStyle
	case "error", "failed", "validation failed":
		return tui.CLIErrorStyle
	case "pending", "not enabled":
		return tui.CLIMutedStyle
	default:
		return tui.CLIWarningStyle
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
