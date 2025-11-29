package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/happytaoer/cli_kanban/internal/model"
)

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#A78BFA")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorDanger    = lipgloss.Color("#EF4444")
	colorMuted     = lipgloss.Color("#6B7280")
	colorBorder    = lipgloss.Color("#374151")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			Width(30)

	columnActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2).
				Width(30)

	columnTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				MarginBottom(1)

	taskStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	taskActiveStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1).
			Background(colorPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Width(60)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	statsStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginBottom(1)
)

// View renders the TUI
func (m Model) View() string {
	switch m.viewMode {
	case ViewModeAddTask:
		return m.viewAddTask()
	case ViewModeEditTask:
		return m.viewEditTask()
	case ViewModeHelp:
		return m.viewHelp()
	default:
		return m.viewBoard()
	}
}

// viewBoard renders the kanban board
func (m Model) viewBoard() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("📋 Kanban Board")
	b.WriteString(title)
	b.WriteString("\n")

	// Statistics
	stats := m.renderStats()
	b.WriteString(stats)
	b.WriteString("\n\n")

	// Columns
	columns := make([]string, len(m.columns))
	for i, col := range m.columns {
		columns[i] = m.renderColumn(i, col)
	}

	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	b.WriteString(columnsView)
	b.WriteString("\n\n")

	// Help text
	help := helpStyle.Render("← → / h l: Navigate columns | ↑ ↓ / j k: Navigate tasks | a: Add | e: Edit | d: Delete | m: Move | ?: Help | q: Quit")
	b.WriteString(help)

	// Error message
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return b.String()
}

// renderStats renders the statistics bar
func (m Model) renderStats() string {
	var parts []string
	for _, col := range m.columns {
		count := len(col.Tasks)
		parts = append(parts, fmt.Sprintf("%s: %d", col.Name, count))
	}
	statsText := strings.Join(parts, " | ")
	return statsStyle.Render(statsText)
}

// renderColumn renders a single column
func (m Model) renderColumn(index int, col model.Column) string {
	var b strings.Builder

	// Column title
	title := columnTitleStyle.Render(col.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Tasks
	if len(col.Tasks) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true).
			Render("No tasks")
		b.WriteString(emptyMsg)
	} else {
		for i, task := range col.Tasks {
			isActive := index == m.currentColumn && i == m.currentTask
			taskView := m.renderTask(task, isActive)
			b.WriteString(taskView)
			b.WriteString("\n")
		}
	}

	// Apply column style
	content := b.String()
	if index == m.currentColumn {
		return columnActiveStyle.Render(content)
	}
	return columnStyle.Render(content)
}

// renderTask renders a single task
func (m Model) renderTask(task model.Task, isActive bool) string {
	text := fmt.Sprintf("• %s", task.Title)
	if isActive {
		return taskActiveStyle.Render(text)
	}
	return taskStyle.Render(text)
}

// viewAddTask renders the add task view
func (m Model) viewAddTask() string {
	var b strings.Builder

	title := titleStyle.Render("➕ Add New Task")
	b.WriteString(title)
	b.WriteString("\n\n")

	col := m.columns[m.currentColumn]
	info := fmt.Sprintf("Adding to column: %s", col.Name)
	b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render(info))
	b.WriteString("\n\n")

	input := inputStyle.Render(m.textInput.View())
	b.WriteString(input)
	b.WriteString("\n\n")

	help := helpStyle.Render("Enter: Save | Esc: Cancel")
	b.WriteString(help)

	return b.String()
}

// viewEditTask renders the edit task view
func (m Model) viewEditTask() string {
	var b strings.Builder

	title := titleStyle.Render("✏️  Edit Task")
	b.WriteString(title)
	b.WriteString("\n\n")

	input := inputStyle.Render(m.textInput.View())
	b.WriteString(input)
	b.WriteString("\n\n")

	help := helpStyle.Render("Enter: Save | Esc: Cancel")
	b.WriteString(help)

	return b.String()
}

// viewHelp renders the help view
func (m Model) viewHelp() string {
	var b strings.Builder

	title := titleStyle.Render("❓ Help")
	b.WriteString(title)
	b.WriteString("\n\n")

	helpText := `Navigation:
  ← → or h l    Move between columns
  ↑ ↓ or j k    Move between tasks

Actions:
  a             Add new task to current column
  e or Enter    Edit selected task
  d or Delete   Delete selected task
  m             Move task to next column

Other:
  ?             Show this help
  q or Ctrl+C   Quit application
  Esc           Cancel current action or quit
`

	b.WriteString(helpText)
	b.WriteString("\n")

	help := helpStyle.Render("Press any key to return to board...")
	b.WriteString(help)

	return b.String()
}
