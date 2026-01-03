package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshw/dms-manager/internal/tui"
	"github.com/joshw/dms-manager/pkg/dms"
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

	// If args don't look like ARNs, try to find matching tasks by name
	taskARNs := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "arn:") {
			taskARNs = append(taskARNs, arg)
		} else {
			// Need to list tasks and find by name
			tasks, err := client.ListTasks(ctx)
			if err != nil {
				exitWithError(err)
			}

			found := false
			for _, task := range tasks {
				if task.Name == arg {
					taskARNs = append(taskARNs, task.ARN)
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("%s task '%s' not found\n", tui.CLIWarningStyle.Render("Warning:"), arg)
			}
		}
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
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("ARN:"), tui.CLIMutedStyle.Render(task.ARN))
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Status:"), getDescribeStatusStyle(task.Status).Render(task.Status))
	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Migration Type:"), tui.CLIValueStyle.Render(task.MigrationType))

	// Endpoints section
	fmt.Println("\n" + tui.CLIHighlightStyle.Render("Endpoints:"))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Replication Instance:"), tui.CLIMutedStyle.Render(task.ReplicationInstanceARN))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Source:"), tui.CLIMutedStyle.Render(task.SourceEndpointARN))
	fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Target:"), tui.CLIMutedStyle.Render(task.TargetEndpointARN))

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
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loaded:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d", stats.TablesLoaded)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loading:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d", stats.TablesLoading)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Queued:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d", stats.TablesQueued)))
		fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Errored:"), getDescribeErrorCountStyle(stats.TablesErrored).Render(fmt.Sprintf("%d", stats.TablesErrored)))

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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Colored headers
	headers := []string{"SCHEMA", "TABLE", "INSERTS", "UPDATES", "DELETES", "DDLS", "ROWS", "STATE"}
	var coloredHeaders []string
	for _, h := range headers {
		coloredHeaders = append(coloredHeaders, tui.CLIHeaderStyle.Render(h))
	}
	fmt.Fprintln(w, "  "+strings.Join(coloredHeaders, "\t"))

	// Separator line
	separators := []string{"──────", "─────", "───────", "───────", "───────", "────", "────", "─────"}
	var coloredSeps []string
	for _, s := range separators {
		coloredSeps = append(coloredSeps, tui.CLIMutedStyle.Render(s))
	}
	fmt.Fprintln(w, "  "+strings.Join(coloredSeps, "\t"))

	for _, s := range stats {
		stateStyle := getValidationStateStyle(s.ValidationState)
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			tui.CLIValueStyle.Render(s.SchemaName),
			tui.CLIPrimaryStyle.Render(s.TableName),
			tui.CLINumberStyle.Render(fmt.Sprintf("%d", s.Inserts)),
			tui.CLINumberStyle.Render(fmt.Sprintf("%d", s.Updates)),
			tui.CLINumberStyle.Render(fmt.Sprintf("%d", s.Deletes)),
			tui.CLINumberStyle.Render(fmt.Sprintf("%d", s.Ddls)),
			tui.CLINumberStyle.Render(fmt.Sprintf("%d", s.FullLoadRows)),
			stateStyle.Render(s.ValidationState),
		)
	}
	w.Flush()
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
