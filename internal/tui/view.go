package tui

import (
	"fmt"
	"strings"
	"time"

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
	case ViewModeEditDue:
		return m.viewEditDue()
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

	// Footer with help text or search input (fixed at bottom)
	var footerContent string
	helpWidth := m.width
	if helpWidth <= 0 {
		helpWidth = 80
	}

	if m.viewMode == ViewModeSearch {
		// Show search input in footer
		searchLabel := lipgloss.NewStyle().Bold(true).Render("Search: ")
		footerContent = searchLabel + m.searchInput.View()
	} else if m.searchQuery != "" {
		// Show active search filter
		searchInfo := lipgloss.NewStyle().Render(fmt.Sprintf("Filter: \"%s\"", m.searchQuery))
		helpText := "/ : Search | Esc: Clear filter | F5: Refresh | ← → : Navigate | a: Add | e: Edit | ?: Help | q: Quit"
		footerContent = searchInfo + "  |  " + helpText
	} else {
		// Normal help text
		footerContent = "← → : Navigate | a: Add | e: Edit | i: Desc | t: Tags | u: Due | d: Del | m: Move | / : Search | F5: Refresh | ?: Help | q: Quit"
	}

	helpContent := lipgloss.PlaceHorizontal(helpWidth, lipgloss.Left, footerContent)
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

	// Filter tasks by search query
	var filteredTasks []model.Task
	for _, task := range col.Tasks {
		if m.matchesSearch(task) {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Column title with scroll indicator
	totalTasks := len(filteredTasks)
	offset := m.scrollOffsets[index]
	if offset >= totalTasks {
		offset = 0
	}
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
			task := filteredTasks[i]
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

// wrapText wraps text at maxWidth using character-based breaking (like HTML break-all)
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}
	var result strings.Builder
	lineWidth := 0
	for _, r := range text {
		charWidth := runeWidth(r)
		if lineWidth+charWidth > maxWidth {
			result.WriteRune('\n')
			lineWidth = 0
		}
		result.WriteRune(r)
		lineWidth += charWidth
	}
	return result.String()
}

// runeWidth returns the display width of a rune (CJK chars are 2, others are 1)
func runeWidth(r rune) int {
	// CJK characters typically take 2 columns
	if r >= 0x4E00 && r <= 0x9FFF || // CJK Unified Ideographs
		r >= 0x3400 && r <= 0x4DBF || // CJK Unified Ideographs Extension A
		r >= 0xFF00 && r <= 0xFFEF { // Fullwidth Forms
		return 2
	}
	return 1
}

// renderTask renders a single task
func (m Model) renderTask(task model.Task, isActive bool) string {
	var b strings.Builder

	// Get max width for text wrapping (account for padding)
	maxWidth := taskStyle.GetWidth() - 2 // subtract padding
	if maxWidth <= 0 {
		maxWidth = 22
	}

	// Wrap title text using character-based breaking
	wrappedTitle := wrapText(task.Title, maxWidth)
	b.WriteString(wrappedTitle)

	// Render due date if present (below title)
	if task.Due != nil {
		dueStr := task.Due.Format("2006-01-02")
		dueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		b.WriteString("\n")
		b.WriteString(dueStyle.Render("📅 " + dueStr))
	}

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

// matchesSearch checks if a task matches the current search query
func (m Model) matchesSearch(task model.Task) bool {
	if m.searchQuery == "" {
		return true
	}

	query := m.searchQuery

	// Check for title: prefix (title-only search)
	if strings.HasPrefix(query, "title:") {
		titleQuery := strings.TrimPrefix(query, "title:")
		if titleQuery == "" {
			return true
		}
		return strings.Contains(strings.ToLower(task.Title), titleQuery)
	}

	// Check for desc: prefix (description-only search)
	if strings.HasPrefix(query, "desc:") {
		descQuery := strings.TrimPrefix(query, "desc:")
		if descQuery == "" {
			return true
		}
		return strings.Contains(strings.ToLower(task.Description), descQuery)
	}

	// Check for tag: prefix (tag-only search)
	if strings.HasPrefix(query, "tag:") {
		tagQuery := strings.TrimPrefix(query, "tag:")
		if tagQuery == "" {
			return true
		}

		for _, tag := range task.Tags {
			if strings.EqualFold(tag, tagQuery) {
				return true
			}
		}

		return false
	}

	// Check for due: prefix (due date search)
	if strings.HasPrefix(query, "due:") {
		dueQuery := strings.TrimPrefix(query, "due:")
		if dueQuery == "" {
			return true
		}

		// Use local time for date comparisons to respect user's timezone
		now := time.Now()
		loc := now.Location()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

		// Special keywords
		if dueQuery == "none" {
			return task.Due == nil
		}
		if dueQuery == "today" {
			if task.Due == nil {
				return false
			}
			taskDue := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, loc)
			return taskDue.Equal(today)
		}
		if dueQuery == "yesterday" {
			if task.Due == nil {
				return false
			}
			yesterday := today.AddDate(0, 0, -1)
			taskDue := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, loc)
			return taskDue.Equal(yesterday)
		}
		if dueQuery == "tomorrow" {
			if task.Due == nil {
				return false
			}
			tomorrow := today.AddDate(0, 0, 1)
			taskDue := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, loc)
			return taskDue.Equal(tomorrow)
		}
		if dueQuery == "overdue" {
			if task.Due == nil {
				return false
			}
			taskDue := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, loc)
			return taskDue.Before(today)
		}

		// If task has no due date (for comparison operators)
		if task.Due == nil {
			return false
		}

		taskDue := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, loc)

		// Check for comparison operators
		if strings.HasPrefix(dueQuery, "<=") {
			dateStr := strings.TrimPrefix(dueQuery, "<=")
			if queryDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				queryDate = time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, loc)
				return !taskDue.After(queryDate)
			}
			return false
		}
		if strings.HasPrefix(dueQuery, ">=") {
			dateStr := strings.TrimPrefix(dueQuery, ">=")
			if queryDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				queryDate = time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, loc)
				return !taskDue.Before(queryDate)
			}
			return false
		}
		if strings.HasPrefix(dueQuery, "<") {
			dateStr := strings.TrimPrefix(dueQuery, "<")
			if queryDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				queryDate = time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, loc)
				return taskDue.Before(queryDate)
			}
			return false
		}
		if strings.HasPrefix(dueQuery, ">") {
			dateStr := strings.TrimPrefix(dueQuery, ">")
			if queryDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				queryDate = time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, loc)
				return taskDue.After(queryDate)
			}
			return false
		}

		// Exact match
		taskDueStr := task.Due.Format("2006-01-02")
		return taskDueStr == dueQuery
	}

	// General search: title, description, tags
	// Search in title
	if strings.Contains(strings.ToLower(task.Title), query) {
		return true
	}

	// Search in description
	if strings.Contains(strings.ToLower(task.Description), query) {
		return true
	}

	// Search in tags
	for _, tag := range task.Tags {
		if strings.EqualFold(tag, query) {
			return true
		}
	}

	return false
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

// viewEditDue renders the edit due date view
func (m Model) viewEditDue() string {
	var b strings.Builder

	title := titleStyle.Render("📅 Edit Due Date")
	b.WriteString(title)
	b.WriteString("\n\n")

	task := m.getCurrentTask()
	if task != nil {
		info := fmt.Sprintf("Task: %s", task.Title)
		b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render(info))
		b.WriteString("\n\n")

		if task.Due != nil {
			currentDue := fmt.Sprintf("Current due: %s", task.Due.Format("2006-01-02"))
			b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(currentDue))
			b.WriteString("\n\n")
		}
	}

	hint := lipgloss.NewStyle().Foreground(colorMuted).Render("Format: YYYY-MM-DD (leave empty to clear)")
	b.WriteString(hint)
	b.WriteString("\n\n")

	input := inputStyle.Render(m.dueInput.View())
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
  u             Edit selected task due date
  d or Delete   Delete selected task
  m             Move task to next column

Search:
  /             Open search input
  Enter         Apply search filter
  Esc           Clear search filter (when active)
  
  Search syntax:
    keyword      Search in title, description and tags
    title:text   Search only in title
    desc:text    Search only in description
    tag:name     Search only in tags (exact match)
    due:YYYY-MM-DD   Exact due date match
    due:<YYYY-MM-DD  Due before date
    due:>YYYY-MM-DD  Due after date
    due:<=YYYY-MM-DD Due on or before date
    due:>=YYYY-MM-DD Due on or after date
    due:today    Due today
    due:yesterday Due yesterday
    due:tomorrow Due tomorrow
    due:overdue  Past due date
    due:none     No due date set

Other:
  F5            Refresh board
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
