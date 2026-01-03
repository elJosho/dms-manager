package tui

import "github.com/charmbracelet/lipgloss"

// lipglossStyle is a type alias for lipgloss.Style
type lipglossStyle = lipgloss.Style

var (
	// Color scheme
	primaryColor   = lipgloss.Color("86")  // Cyan
	secondaryColor = lipgloss.Color("212") // Pink
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red
	warningColor   = lipgloss.Color("220") // Yellow
	mutedColor     = lipgloss.Color("241") // Gray
	accentColor    = lipgloss.Color("147") // Light purple
	highlightColor = lipgloss.Color("215") // Orange
	infoBlueColor  = lipgloss.Color("39")  // Bright blue

	// Common styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle()

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	statusOtherStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	selectedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor).
				Padding(1, 2).
				Bold(true)

	// Label and value styles for details view
	labelStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray/white

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				MarginTop(1)

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(infoBlueColor).
				Bold(true)

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")) // Light green

	arnStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	checkmarkStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Additional TUI view styles
	warningTextStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	mutedTextStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	mutedCheckboxStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	selectedCursorStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	errorTextStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// CLI-specific exported styles
	CLIPrimaryStyle   = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	CLISecondaryStyle = lipgloss.NewStyle().Foreground(secondaryColor)
	CLISuccessStyle   = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	CLIErrorStyle     = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	CLIWarningStyle   = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	CLIMutedStyle     = lipgloss.NewStyle().Foreground(mutedColor)
	CLILabelStyle     = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	CLIValueStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	CLIHeaderStyle    = lipgloss.NewStyle().Foreground(infoBlueColor).Bold(true)
	CLIHighlightStyle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true)
	CLINumberStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
)

// GetStatusStyle returns the appropriate style for a task status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "running", "starting", "replicating":
		return statusRunningStyle
	case "stopped", "stopping", "failed":
		return statusStoppedStyle
	default:
		return statusOtherStyle
	}
}
