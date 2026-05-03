package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljosho/dms-manager/internal/tui"
	"github.com/eljosho/dms-manager/pkg/dms"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DMS replication tasks",
	Long:  `List all DMS replication tasks in the selected AWS account and region.`,
	Run:   runList,
}

func init() {
	listCmd.Flags().Bool("stats", false, "Show detailed table statistics for each task")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	client, err := dms.NewClient(ctx, GetProfile(), GetRegion())
	if err != nil {
		exitWithError(fmt.Errorf("failed to create DMS client: %w", err))
	}

	fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Region:"), tui.CLIPrimaryStyle.Render(client.GetRegion()))
	if client.GetProfile() != "" {
		fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Profile:"), tui.CLISecondaryStyle.Render(client.GetProfile()))
	}
	fmt.Println()

	tasks, err := client.ListTasks(ctx)
	if err != nil {
		exitWithError(err)
	}

	if len(tasks) == 0 {
		fmt.Println(tui.CLIWarningStyle.Render("No DMS replication tasks found."))
		return
	}

	showStats, _ := cmd.Flags().GetBool("stats")

	if showStats {
		// Print tasks with detailed statistics
		for i, task := range tasks {
			if i > 0 {
				fmt.Println(tui.CLIMutedStyle.Render(strings.Repeat("─", 80)))
			}

			fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Task:"), tui.CLIPrimaryStyle.Render(task.Name))
			fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Status:"), getListStatusStyle(task.Status).Render(task.Status))
			fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("Type:"), tui.CLIValueStyle.Render(task.MigrationType))
			fmt.Printf("%s %s\n", tui.CLILabelStyle.Render("ARN:"), tui.CLIMutedStyle.Render(dms.TruncateARN(task.ARN, 60)))

			if task.ReplicationTaskStats != nil {
				stats := task.ReplicationTaskStats
				fmt.Println("\n" + tui.CLIHighlightStyle.Render("Statistics:"))
				fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Full Load Progress:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d%%", stats.FullLoadProgressPercent)))
				fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loaded:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesLoaded)))
				fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Loading:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesLoading)))
				fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Queued:"), tui.CLINumberStyle.Render(dms.FormatNumber(stats.TablesQueued)))
				fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Tables Errored:"), getListErrorCountStyle(stats.TablesErrored).Render(dms.FormatNumber(stats.TablesErrored)))

				if stats.ElapsedTimeMillis > 0 {
					fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Elapsed Time:"), tui.CLIValueStyle.Render(dms.FormatElapsedTime(stats.ElapsedTimeMillis)))
				}

				if stats.StopReason != "" {
					fmt.Printf("  %s %s\n", tui.CLILabelStyle.Render("Stop Reason:"), tui.CLIWarningStyle.Render(stats.StopReason))
				}
			}
			fmt.Println()
		}
	} else {
		// Print tasks in a table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, tui.CLIHeaderStyle.Render("NAME")+"\t"+tui.CLIHeaderStyle.Render("STATUS")+"\t"+tui.CLIHeaderStyle.Render("%")+"\t"+tui.CLIHeaderStyle.Render("TABLES")+"\t"+tui.CLIHeaderStyle.Render("ELAPSED")+"\t"+tui.CLIHeaderStyle.Render("TYPE")+"\t"+tui.CLIHeaderStyle.Render("ARN"))
		fmt.Fprintln(w, tui.CLIMutedStyle.Render("────")+"\t"+tui.CLIMutedStyle.Render("──────")+"\t"+tui.CLIMutedStyle.Render("─")+"\t"+tui.CLIMutedStyle.Render("──────")+"\t"+tui.CLIMutedStyle.Render("───────")+"\t"+tui.CLIMutedStyle.Render("────")+"\t"+tui.CLIMutedStyle.Render("───"))

		for _, task := range tasks {
			progress := "-"
			tablesStr := "-"
			elapsedStr := "-"
			if task.ReplicationTaskStats != nil {
				progress = fmt.Sprintf("%d%%", task.ReplicationTaskStats.FullLoadProgressPercent)
				totalTables := task.ReplicationTaskStats.TablesLoaded + task.ReplicationTaskStats.TablesLoading + task.ReplicationTaskStats.TablesQueued + task.ReplicationTaskStats.TablesErrored
				tablesStr = dms.FormatNumber(totalTables)
				elapsedStr = dms.FormatElapsedTime(task.ReplicationTaskStats.ElapsedTimeMillis)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				tui.CLIPrimaryStyle.Render(task.Name),
				getListStatusStyle(task.Status).Render(task.Status),
				tui.CLINumberStyle.Render(progress),
				tui.CLINumberStyle.Render(tablesStr),
				tui.CLIValueStyle.Render(elapsedStr),
				tui.CLIValueStyle.Render(task.MigrationType),
				tui.CLIMutedStyle.Render(dms.TruncateARN(task.ARN, 40)),
			)
		}

		w.Flush()
	}

	fmt.Printf("\n%s %s\n", tui.CLILabelStyle.Render("Total tasks:"), tui.CLINumberStyle.Render(fmt.Sprintf("%d", len(tasks))))
}

// getListStatusStyle returns the appropriate style for a task status
func getListStatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "running", "starting", "replicating":
		return tui.CLISuccessStyle
	case "stopped", "stopping", "failed":
		return tui.CLIErrorStyle
	default:
		return tui.CLIWarningStyle
	}
}

// getListErrorCountStyle returns red for non-zero error counts
func getListErrorCountStyle(count int32) lipgloss.Style {
	if count > 0 {
		return tui.CLIErrorStyle
	}
	return tui.CLINumberStyle
}
