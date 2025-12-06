package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/happytaoer/cli_kanban/internal/model"
)

var (
	// Colors
	colorPrimary    = lipgloss.Color("#7C3AED")
	colorSecondary  = lipgloss.Color("#A78BFA")
	colorInProgress = lipgloss.Color("#3B82F6")
	colorSuccess    = lipgloss.Color("#10B981")
	colorDanger     = lipgloss.Color("#EF4444")
	colorMuted      = lipgloss.Color("#6B7280")
	colorBorder     = lipgloss.Color("#374151")

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

	columnTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				MarginBottom(1)

	taskStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1).
			Width(26)

	taskActiveStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1).
			Width(26).
			Background(colorPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	footerStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Foreground(colorMuted).
			PaddingTop(1)

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
	case ViewModeEditDescription:
		return m.viewEditDescription()
	case ViewModeEditTags:
		return m.viewEditTags()
	case ViewModeConfirmDelete:
		return m.viewConfirmDelete()
	case ViewModeHelp:
		return m.viewHelp()
	default:
		return m.viewBoard()
	}
}

// viewBoard renders the kanban board
func (m Model) viewBoard() string {
	// Header: Title + Statistics on same line
	title := titleStyle.Render("📋 Kanban Board")
	stats := m.renderStats()
	headerWidth := m.width
	if headerWidth <= 0 {
		headerWidth = 80
	}
	// Place title on left, stats on right
	spacerWidth := headerWidth - lipgloss.Width(title) - lipgloss.Width(stats)
	if spacerWidth < 0 {
		spacerWidth = 1
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		lipgloss.NewStyle().Width(spacerWidth).Render(""),
		stats,
	)

	// Columns content for viewport
	columns := make([]string, len(m.columns))
	for i, col := range m.columns {
		columns[i] = m.renderColumn(i, col)
	}
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Error message appended to columns if present
	if m.err != nil {
		columnsView += "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Set viewport content and render
	m.viewport.SetContent(columnsView)

	// Footer with help text (fixed at bottom)
	helpText := "← → / h l: Navigate | a: Add | e: Edit | i: Desc | t: Tags | d: Del | m: Move | ?: Help | q: Quit"
	helpWidth := m.width
	if helpWidth <= 0 {
		helpWidth = lipgloss.Width(helpText)
	}
	helpContent := lipgloss.PlaceHorizontal(helpWidth, lipgloss.Left, helpText)
	footer := footerStyle.Width(helpWidth).Render(helpContent)

	// Combine: header + viewport + footer
	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// renderStats renders the statistics bar
func (m Model) renderStats() string {
	var parts []string
	for _, col := range m.columns {
		count := len(col.Tasks)
		parts = append(parts, fmt.Sprintf("%s: %d", col.Name, count))
	}
	statsText := strings.Join(parts, " | ")
	if !m.currentTime.IsZero() {
		statsText = fmt.Sprintf("%s | 🕒 %s", statsText, m.currentTime.Format("2006-01-02 15:04:05"))
	}
	return statsStyle.Render(statsText)
}

// renderColumn renders a single column
func (m Model) renderColumn(index int, col model.Column) string {
	var b strings.Builder

	// Column title with scroll indicator
	totalTasks := len(col.Tasks)
	offset := m.scrollOffsets[index]
	titleStyle := columnTitleStyle
	switch col.Status {
	case model.StatusInProgress:
		titleStyle = titleStyle.Copy().Foreground(colorInProgress)
	case model.StatusDone:
		titleStyle = titleStyle.Copy().Foreground(colorSuccess)
	default:
		titleStyle = titleStyle.Copy().Foreground(colorMuted)
	}
	title := titleStyle.Render(col.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Scroll up indicator
	if offset > 0 {
		scrollUp := lipgloss.NewStyle().Foreground(colorMuted).Render("  ▲ more above")
		b.WriteString(scrollUp)
		b.WriteString("\n")
	}

	// Tasks (only visible range)
	if totalTasks == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true).
			Render("No tasks")
		b.WriteString(emptyMsg)
	} else {
		endIndex := offset + maxVisibleTasks
		if endIndex > totalTasks {
			endIndex = totalTasks
		}
		for i := offset; i < endIndex; i++ {
			task := col.Tasks[i]
			isActive := index == m.currentColumn && i == m.currentTask
			taskView := m.renderTask(task, isActive)
			b.WriteString(taskView)
			b.WriteString("\n")
		}
	}

	// Scroll down indicator
	if offset+maxVisibleTasks < totalTasks {
		scrollDown := lipgloss.NewStyle().Foreground(colorMuted).Render("  ▼ more below")
		b.WriteString(scrollDown)
	}

	// Apply column style with status-specific colors
	content := b.String()
	style := columnStyle.Copy()
	switch col.Status {
	case model.StatusInProgress:
		style = style.BorderForeground(colorInProgress)
	case model.StatusDone:
		style = style.BorderForeground(colorSuccess)
	default:
		style = style.BorderForeground(colorMuted)
	}
	if index == m.currentColumn {
		style = style.Copy().Bold(true)
	}
	return style.Render(content)
}

// renderTask renders a single task
func (m Model) renderTask(task model.Task, isActive bool) string {
	var b strings.Builder
	b.WriteString(task.Title)

	// Render tags if present
	if len(task.Tags) > 0 {
		b.WriteString("\n")
		lineWidth := 0
		maxWidth := taskStyle.GetWidth()
		if maxWidth == 0 {
			maxWidth = 24
		}
		for _, tag := range task.Tags {
			tagStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(getTagColor(tag)).
				Padding(0, 1)
			rendered := tagStyle.Render(tag)
			tagWidth := lipgloss.Width(rendered)
			space := 0
			if lineWidth > 0 {
				space = 1
			}
			if lineWidth+space+tagWidth > maxWidth {
				b.WriteString("\n")
				lineWidth = 0
				space = 0
			}
			if space == 1 {
				b.WriteString(" ")
				lineWidth++
			}
			b.WriteString(rendered)
			lineWidth += tagWidth
		}
	}

	text := b.String()
	if isActive {
		return taskActiveStyle.Render(text)
	}
	return taskStyle.Render(text)
}

// getTagColor returns a color based on tag name hash
func getTagColor(tag string) lipgloss.Color {
	colors := []lipgloss.Color{
		lipgloss.Color("#EF4444"), // red
		lipgloss.Color("#F59E0B"), // orange
		lipgloss.Color("#10B981"), // green
		lipgloss.Color("#3B82F6"), // blue
		lipgloss.Color("#8B5CF6"), // purple
		lipgloss.Color("#EC4899"), // pink
	}
	hash := 0
	for _, c := range tag {
		hash += int(c)
	}
	return colors[hash%len(colors)]
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

// viewEditDescription renders the edit description view
func (m Model) viewEditDescription() string {
	var b strings.Builder

	title := titleStyle.Render("📝 Edit Task Description")
	b.WriteString(title)
	b.WriteString("\n\n")

	task := m.getCurrentTask()
	if task != nil {
		info := fmt.Sprintf("Task: %s", task.Title)
		b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render(info))
		b.WriteString("\n\n")
	}

	textAreaView := m.textArea.View()
	b.WriteString(textAreaView)
	b.WriteString("\n\n")

	help := helpStyle.Render("Ctrl+S: Save | Esc: Cancel")
	b.WriteString(help)

	return b.String()
}

// viewEditTags renders the edit tags view
func (m Model) viewEditTags() string {
	var b strings.Builder

	title := titleStyle.Render("🏷️  Edit Tags")
	b.WriteString(title)
	b.WriteString("\n\n")

	task := m.getCurrentTask()
	if task != nil {
		info := fmt.Sprintf("Task: %s", task.Title)
		b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render(info))
		b.WriteString("\n\n")
	}

	hint := lipgloss.NewStyle().Foreground(colorMuted).Render("Separate tags with commas (e.g., bug, urgent, feature)")
	b.WriteString(hint)
	b.WriteString("\n\n")

	input := inputStyle.Render(m.textInput.View())
	b.WriteString(input)
	b.WriteString("\n\n")

	help := helpStyle.Render("Enter: Save | Esc: Cancel")
	b.WriteString(help)

	return b.String()
}

// viewConfirmDelete renders the delete confirmation view
func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	title := titleStyle.Render("⚠️  Confirm Delete")
	b.WriteString(title)
	b.WriteString("\n\n")

	task := m.getCurrentTask()
	if task != nil {
		warning := lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true).
			Render(fmt.Sprintf("Are you sure you want to delete this task?\n\n\"%s\"", task.Title))
		b.WriteString(warning)
		b.WriteString("\n\n")
	}

	help := helpStyle.Render("y: Yes, delete | n/Esc: Cancel")
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
  e or Enter    Edit selected task title
  i             Edit selected task description
  t             Edit selected task tags
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
