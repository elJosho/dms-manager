package tui

import (
	"fmt"
	"strings"

	"github.com/joshw/dms-manager/pkg/dms"
)

// View renders the current view
func (m Model) View() string {
	switch m.state {
	case viewLoading:
		return m.renderLoading()
	case viewError:
		return m.renderError()
	case viewTaskList:
		return m.renderTaskList()
	case viewTaskDetails:
		return m.renderTaskDetails()
	case viewTableStats:
		return m.renderTableStats()
	default:
		return "Unknown state"
	}
}

func (m Model) renderLoading() string {
	return fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), titleStyle.Render("Loading DMS tasks..."))
}

func (m Model) renderError() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Error"))
	sb.WriteString("\n\n")
	sb.WriteString(errorStyle.Render(fmt.Sprintf("Failed to load tasks: %v", m.err)))
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("Press 'r' to retry or 'q' to quit"))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) renderTaskList() string {
	var sb strings.Builder

	// Header
	title := fmt.Sprintf("AWS DMS Tasks - %s", m.client.GetRegion())
	if m.client.GetProfile() != "" {
		title += fmt.Sprintf(" (Profile: %s)", m.client.GetProfile())
	}
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n")

	// Operation message
	if m.operationMsg != "" {
		sb.WriteString(infoStyle.Render(m.operationMsg))
		sb.WriteString("\n\n")
	}

	// Task list
	if len(m.tasks) == 0 {
		sb.WriteString(warningTextStyle.Render("No tasks found."))
		sb.WriteString("\n")
	} else {
		for i, task := range m.tasks {
			checkbox := mutedCheckboxStyle.Render("[ ]")
			if m.selected[i] {
				checkbox = checkmarkStyle.Render("[✓]")
			}

			cursor := "  "
			nameStyle := normalItemStyle
			if i == m.cursor {
				cursor = selectedCursorStyle.Render("→ ")
				nameStyle = selectedItemStyle
			}

			statusStyle := GetStatusStyle(strings.ToLower(task.Status))
			status := statusStyle.Render(task.Status)
			migrationType := mutedTextStyle.Render(fmt.Sprintf("(%s)", task.MigrationType))

			line := fmt.Sprintf("%s%s %s - %s %s", cursor, checkbox, nameStyle.Render(task.Name), status, migrationType)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderHelp())

	return sb.String()
}

func (m Model) renderTaskDetails() string {
	if m.detailsTaskIdx >= len(m.tasks) {
		return "Invalid task index"
	}

	task := m.tasks[m.detailsTaskIdx]

	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Task Details"))
	sb.WriteString("\n\n")

	// Basic info with colored labels
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Name:"), valueStyle.Render(task.Name)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status:"), GetStatusStyle(strings.ToLower(task.Status)).Render(task.Status)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Type:"), valueStyle.Render(task.MigrationType)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ARN:"), arnStyle.Render(task.ARN)))
	sb.WriteString("\n")

	// Endpoints section
	sb.WriteString(sectionHeaderStyle.Render("Endpoints:"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Source:"), arnStyle.Render(task.SourceEndpointARN)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Target:"), arnStyle.Render(task.TargetEndpointARN)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Instance:"), arnStyle.Render(task.ReplicationInstanceARN)))
	sb.WriteString("\n")

	// Timestamps
	if task.CreatedAt != nil {
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Created:"), valueStyle.Render(task.CreatedAt.Format("2006-01-02 15:04:05"))))
	}
	if task.StartedAt != nil {
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Started:"), valueStyle.Render(task.StartedAt.Format("2006-01-02 15:04:05"))))
	}

	// Statistics
	if task.ReplicationTaskStats != nil {
		stats := task.ReplicationTaskStats
		sb.WriteString("\n")
		sb.WriteString(sectionHeaderStyle.Render("Statistics:"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Progress:"), numberStyle.Render(fmt.Sprintf("%d%%", stats.FullLoadProgressPercent))))

		// Tables summary with conditional coloring for errors
		tablesErroredStyle := numberStyle
		if stats.TablesErrored > 0 {
			tablesErroredStyle = errorStyle
		}

		sb.WriteString(fmt.Sprintf("  %s %s: %s, %s: %s, %s: %s, %s: %s\n",
			labelStyle.Render("Tables -"),
			labelStyle.Render("Loaded"), numberStyle.Render(fmt.Sprintf("%d", stats.TablesLoaded)),
			labelStyle.Render("Loading"), numberStyle.Render(fmt.Sprintf("%d", stats.TablesLoading)),
			labelStyle.Render("Queued"), numberStyle.Render(fmt.Sprintf("%d", stats.TablesQueued)),
			labelStyle.Render("Errored"), tablesErroredStyle.Render(fmt.Sprintf("%d", stats.TablesErrored))))

		if stats.ElapsedTimeMillis > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Elapsed:"), valueStyle.Render(dms.FormatElapsedTime(stats.ElapsedTimeMillis))))
		}

		if stats.StopReason != "" {
			sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Stop Reason:"), statusOtherStyle.Render(stats.StopReason)))
		}
	}

	// Extended stats - Table Mappings
	if m.showExtendedStats && task.TableMappings != "" {
		sb.WriteString("\n")
		sb.WriteString(sectionHeaderStyle.Render("Table Mappings:"))
		sb.WriteString("\n")
		sb.WriteString(mutedTextStyle.Render(task.TableMappings))
		sb.WriteString("\n")
	}

	// Error info
	if task.LastFailureMessage != "" {
		sb.WriteString("\n")
		sb.WriteString(errorStyle.Render("Last Failure:"))
		sb.WriteString("\n")
		sb.WriteString(errorTextStyle.Render(task.LastFailureMessage))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Build help string parts
	var parts []string

	if task.TableMappings != "" {
		extendedStatus := mutedTextStyle.Render("off")
		if m.showExtendedStats {
			extendedStatus = statusRunningStyle.Render("on")
		}
		parts = append(parts, fmt.Sprintf("[t] extended stats: %s", extendedStatus))
	}

	parts = append(parts, "[T] table stats")
	parts = append(parts, "[ESC] back")

	helpText := fmt.Sprintf("Press %s", strings.Join(parts, " • "))

	sb.WriteString(helpStyle.Render(helpText))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) renderTableStats() string {
	if m.detailsTaskIdx >= len(m.tasks) {
		return "Invalid task index"
	}

	task := m.tasks[m.detailsTaskIdx]

	var sb strings.Builder

	sb.WriteString(titleStyle.Render(fmt.Sprintf("Table Statistics - %s", task.Name)))
	sb.WriteString("\n\n")

	if len(m.tableStats) == 0 {
		sb.WriteString(warningTextStyle.Render("No table statistics available."))
		sb.WriteString("\n")
	} else {
		// Table header with colors
		header := fmt.Sprintf("%-15s %-15s %-8s %-8s %-8s %-8s %-8s %-10s",
			"SCHEMA", "TABLE", "INSERTS", "UPDATES", "DELETES", "DDLS", "ROWS", "STATE")
		sb.WriteString(tableHeaderStyle.Render(header))
		sb.WriteString("\n")

		separator := fmt.Sprintf("%-15s %-15s %-8s %-8s %-8s %-8s %-8s %-10s",
			"───────────────", "───────────────", "────────", "────────", "────────", "────────", "────────", "──────────")
		sb.WriteString(mutedTextStyle.Render(separator))
		sb.WriteString("\n")

		// Table rows with colored values
		for _, s := range m.tableStats {
			stateStyle := getTableValidationStyle(s.ValidationState)

			sb.WriteString(fmt.Sprintf("%-15s %-15s %s %s %s %s %s %s\n",
				valueStyle.Render(truncateString(s.SchemaName, 15)),
				valueStyle.Render(truncateString(s.TableName, 15)),
				numberStyle.Render(fmt.Sprintf("%-8d", s.Inserts)),
				numberStyle.Render(fmt.Sprintf("%-8d", s.Updates)),
				numberStyle.Render(fmt.Sprintf("%-8d", s.Deletes)),
				numberStyle.Render(fmt.Sprintf("%-8d", s.Ddls)),
				numberStyle.Render(fmt.Sprintf("%-8d", s.FullLoadRows)),
				stateStyle.Render(fmt.Sprintf("%-10s", s.ValidationState)),
			))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("Press [ESC] to go back"))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) renderHelp() string {
	autoRefreshStatus := mutedTextStyle.Render("off")
	if m.autoRefresh {
		autoRefreshStatus = statusRunningStyle.Render("on")
	}

	helpLine1 := []string{
		"[↑/k] up",
		"[↓/j] down",
		"[space] select",
		"[enter] details",
		"[s] start",
		"[x] stop",
		"[r] resume",
		"[l] reload",
	}

	helpLine2 := []string{
		"[c] clear",
		"[f] refresh",
		fmt.Sprintf("[a] auto-refresh: %s", autoRefreshStatus),
		"[q] quit",
	}

	return helpStyle.Render(strings.Join(helpLine1, " • ") + "\n" + strings.Join(helpLine2, " • "))
}

// getTableValidationStyle returns color based on validation state
func getTableValidationStyle(state string) lipglossStyle {
	switch strings.ToLower(state) {
	case "validated", "table validated":
		return statusRunningStyle
	case "error", "failed", "validation failed":
		return errorStyle
	case "pending", "not enabled":
		return mutedTextStyle
	default:
		return statusOtherStyle
	}
}

func truncateARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return arn
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
