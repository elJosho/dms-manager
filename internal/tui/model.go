package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshw/dms-manager/pkg/dms"
)

// View states
type viewState int

const (
	viewTaskList viewState = iota
	viewTaskDetails
	viewTableStats
	viewLoading
	viewError
)

// Model holds the state for the TUI
type Model struct {
	client            *dms.Client
	tasks             []dms.Task
	tableStats        []dms.TableStatistic
	cursor            int
	selected          map[int]bool
	state             viewState
	err               error
	spinner           spinner.Model
	width             int
	height            int
	detailsTaskIdx    int
	operationMsg      string
	autoRefresh       bool
	showExtendedStats bool
}

// NewModel creates a new TUI model
func NewModel(client *dms.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = infoStyle

	return Model{
		client:      client,
		selected:    make(map[int]bool),
		state:       viewLoading,
		spinner:     s,
		autoRefresh: true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		LoadTasksCmd(m.client),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tasksLoadedMsg:
		if msg.err != nil {
			m.state = viewError
			m.err = msg.err
			return m, nil
		}
		m.tasks = msg.tasks
		if m.state == viewLoading {
			m.state = viewTaskList
		}
		if m.autoRefresh {
			return m, TickCmd()
		}
		if m.autoRefresh {
			return m, TickCmd()
		}
		return m, nil

	case tableStatsLoadedMsg:
		if msg.err != nil {
			m.state = viewError
			m.err = msg.err
			return m, nil
		}
		m.tableStats = msg.stats
		m.state = viewTableStats
		return m, nil

	case taskOperationCompleteMsg:
		// Show results and reload tasks
		m.operationMsg = formatOperationResults(msg.results)
		return m, LoadTasksCmd(m.client)

	case tickMsg:
		if m.autoRefresh && m.state == viewTaskList {
			return m, tea.Batch(LoadTasksCmd(m.client), TickCmd())
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case errorMsg:
		m.state = viewError
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "f":
		// Refresh task list (changed from 'r' to avoid conflict with resume)
		m.operationMsg = ""
		return m, LoadTasksCmd(m.client)

	case "a":
		// Toggle auto-refresh
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			return m, TickCmd()
		}
		return m, nil
	}

	switch m.state {
	case viewTaskList:
		return m.handleTaskListKeys(msg)
	case viewTaskDetails:
		return m.handleTaskDetailsKeys(msg)
	case viewTableStats:
		return m.handleTableStatsKeys(msg)
	}

	return m, nil
}

func (m Model) handleTaskListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}

	case " ":
		// Toggle selection
		if m.selected[m.cursor] {
			delete(m.selected, m.cursor)
		} else {
			m.selected[m.cursor] = true
		}

	case "enter":
		// View task details
		m.detailsTaskIdx = m.cursor
		m.state = viewTaskDetails

	case "s":
		// Start selected tasks
		arns := m.getSelectedARNs()
		if len(arns) > 0 {
			m.operationMsg = "Starting tasks..."
			return m, StartTasksCmd(m.client, arns)
		}

	case "x":
		// Stop selected tasks
		arns := m.getSelectedARNs()
		if len(arns) > 0 {
			m.operationMsg = "Stopping tasks..."
			return m, StopTasksCmd(m.client, arns)
		}

	case "r":
		// Resume selected tasks
		arns := m.getSelectedARNs()
		if len(arns) > 0 {
			m.operationMsg = "Resuming tasks..."
			return m, ResumeTasksCmd(m.client, arns)
		}

	case "l":
		// Reload selected tasks
		arns := m.getSelectedARNs()
		if len(arns) > 0 {
			m.operationMsg = "Reloading tasks..."
			return m, ReloadTasksCmd(m.client, arns)
		}

	case "c":
		// Clear selection
		m.selected = make(map[int]bool)
	}

	return m, nil
}

func (m Model) handleTaskDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.state = viewTaskList
	case "t":
		// Toggle extended stats view
		m.showExtendedStats = !m.showExtendedStats

	case "T":
		// View table statistics
		m.state = viewLoading
		return m, LoadTableStatsCmd(m.client, m.tasks[m.detailsTaskIdx].ARN)
	}

	return m, nil
}

func (m Model) handleTableStatsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "q":
		m.state = viewTaskDetails
	}
	return m, nil
}

func (m Model) getSelectedARNs() []string {
	if len(m.selected) == 0 {
		// If nothing selected, use cursor position
		if m.cursor < len(m.tasks) {
			return []string{m.tasks[m.cursor].ARN}
		}
		return nil
	}

	arns := make([]string, 0, len(m.selected))
	for idx := range m.selected {
		if idx < len(m.tasks) {
			arns = append(arns, m.tasks[idx].ARN)
		}
	}
	return arns
}

func formatOperationResults(results []dms.TaskOperation) string {
	var sb strings.Builder
	successCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	sb.WriteString(fmt.Sprintf("Completed %d/%d operations successfully", successCount, len(results)))
	return sb.String()
}
