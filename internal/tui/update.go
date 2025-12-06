package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/happytaoer/cli_kanban/internal/model"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clockTickMsg:
		m.currentTime = time.Time(msg)
		return m, clockTickCmd()

	case tasksLoadedMsg:
		m.organizeTasks(msg.tasks)
		return m, nil

	case taskCreatedMsg:
		return m, m.loadTasks()

	case taskUpdatedMsg:
		return m, m.loadTasks()

	case taskDeletedMsg:
		return m, m.loadTasks()

	case descriptionUpdatedMsg:
		return m, m.loadTasks()

	case tagsUpdatedMsg:
		return m, m.loadTasks()

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Handle text input updates
	if m.viewMode == ViewModeAddTask || m.viewMode == ViewModeEditTask || m.viewMode == ViewModeEditTags {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Handle textarea updates
	if m.viewMode == ViewModeEditDescription {
		m.textArea, cmd = m.textArea.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.viewMode == ViewModeBoard {
			return m, tea.Quit
		}
	case "esc":
		if m.viewMode != ViewModeBoard {
			m.viewMode = ViewModeBoard
			m.textInput.SetValue("")
			return m, nil
		}
		return m, tea.Quit
	}

	// Mode-specific keys
	switch m.viewMode {
	case ViewModeBoard:
		return m.handleBoardKeys(msg)
	case ViewModeAddTask:
		return m.handleAddTaskKeys(msg)
	case ViewModeEditTask:
		return m.handleEditTaskKeys(msg)
	case ViewModeEditDescription:
		return m.handleEditDescriptionKeys(msg)
	case ViewModeEditTags:
		return m.handleEditTagsKeys(msg)
	case ViewModeConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ViewModeHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleBoardKeys handles keyboard input in board view mode
func (m Model) handleBoardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.currentColumn > 0 {
			m.currentColumn--
			m.currentTask = 0
		}
		return m, nil

	case "right", "l":
		if m.currentColumn < len(m.columns)-1 {
			m.currentColumn++
			m.currentTask = 0
		}
		return m, nil

	case "up", "k":
		if m.currentTask > 0 {
			m.currentTask--
			m.ensureTaskVisible()
		}
		return m, nil

	case "down", "j":
		col := m.columns[m.currentColumn]
		if m.currentTask < len(col.Tasks)-1 {
			m.currentTask++
			m.ensureTaskVisible()
		}
		return m, nil

	case "a":
		m.viewMode = ViewModeAddTask
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case "e", "enter":
		task := m.getCurrentTask()
		if task != nil {
			m.viewMode = ViewModeEditTask
			m.textInput.SetValue(task.Title)
			m.textInput.Focus()
		}
		return m, nil

	case "d", "delete":
		task := m.getCurrentTask()
		if task != nil {
			m.pendingDeleteID = task.ID
			m.viewMode = ViewModeConfirmDelete
		}
		return m, nil

	case "m":
		task := m.getCurrentTask()
		if task != nil {
			nextColumn := (m.currentColumn + 1) % len(m.columns)
			m.currentColumn = nextColumn
			m.followTaskID = task.ID
			return m, m.moveTask(task, nextColumn)
		}
		return m, nil

	case "i":
		task := m.getCurrentTask()
		if task != nil {
			m.viewMode = ViewModeEditDescription
			m.textArea.SetValue(task.Description)
			m.textArea.Focus()
		}
		return m, nil

	case "t":
		task := m.getCurrentTask()
		if task != nil {
			m.viewMode = ViewModeEditTags
			m.textInput.SetValue(strings.Join(task.Tags, ", "))
			m.textInput.Focus()
		}
		return m, nil

	case "?":
		m.viewMode = ViewModeHelp
		return m, nil
	}

	return m, nil
}

// handleAddTaskKeys handles keyboard input in add task mode
func (m Model) handleAddTaskKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		title := m.textInput.Value()
		if title != "" {
			status := m.columns[m.currentColumn].Status
			m.viewMode = ViewModeBoard
			m.textInput.SetValue("")
			return m, m.createTask(title, status)
		}
		return m, nil

	case "esc":
		m.viewMode = ViewModeBoard
		m.textInput.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleEditTaskKeys handles keyboard input in edit task mode
func (m Model) handleEditTaskKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		title := m.textInput.Value()
		task := m.getCurrentTask()
		if title != "" && task != nil {
			m.viewMode = ViewModeBoard
			m.textInput.SetValue("")
			return m, m.updateTask(task.ID, title, task.Status)
		}
		return m, nil

	case "esc":
		m.viewMode = ViewModeBoard
		m.textInput.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleEditDescriptionKeys handles keyboard input in edit description mode
func (m Model) handleEditDescriptionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		description := m.textArea.Value()
		task := m.getCurrentTask()
		if task != nil {
			m.viewMode = ViewModeBoard
			m.textArea.SetValue("")
			return m, m.updateDescription(task.ID, description)
		}
		return m, nil

	case "esc":
		m.viewMode = ViewModeBoard
		m.textArea.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

// handleEditTagsKeys handles keyboard input in edit tags mode
func (m Model) handleEditTagsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tagsStr := m.textInput.Value()
		task := m.getCurrentTask()
		if task != nil {
			tags := parseTagsInput(tagsStr)
			m.viewMode = ViewModeBoard
			m.textInput.SetValue("")
			return m, m.updateTags(task.ID, tags)
		}
		return m, nil

	case "esc":
		m.viewMode = ViewModeBoard
		m.textInput.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// parseTagsInput parses comma-separated tags input
func parseTagsInput(input string) []string {
	parts := strings.Split(input, ",")
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(strings.ToLower(p))
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// handleConfirmDeleteKeys handles keyboard input in delete confirmation mode
func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		id := m.pendingDeleteID
		m.pendingDeleteID = 0
		m.viewMode = ViewModeBoard
		return m, m.deleteTask(id)

	case "n", "N", "esc":
		m.pendingDeleteID = 0
		m.viewMode = ViewModeBoard
		return m, nil
	}

	return m, nil
}

// handleHelpKeys handles keyboard input in help mode
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.viewMode = ViewModeBoard
	return m, nil
}

// createTask creates a new task
func (m Model) createTask(title string, status model.TaskStatus) tea.Cmd {
	return func() tea.Msg {
		task, err := m.db.CreateTask(title, status)
		if err != nil {
			return errMsg{err}
		}
		return taskCreatedMsg{task}
	}
}

// updateTask updates a task
func (m Model) updateTask(id int64, title string, status model.TaskStatus) tea.Cmd {
	return func() tea.Msg {
		err := m.db.UpdateTask(id, title, status)
		if err != nil {
			return errMsg{err}
		}
		return taskUpdatedMsg{}
	}
}

// deleteTask deletes a task
func (m Model) deleteTask(id int64) tea.Cmd {
	return func() tea.Msg {
		err := m.db.DeleteTask(id)
		if err != nil {
			return errMsg{err}
		}
		return taskDeletedMsg{}
	}
}

// updateDescription updates a task's description
func (m Model) updateDescription(id int64, description string) tea.Cmd {
	return func() tea.Msg {
		err := m.db.UpdateTaskDescription(id, description)
		if err != nil {
			return errMsg{err}
		}
		return descriptionUpdatedMsg{}
	}
}

// updateTags updates a task's tags
func (m Model) updateTags(id int64, tags []string) tea.Cmd {
	return func() tea.Msg {
		err := m.db.UpdateTaskTags(id, tags)
		if err != nil {
			return errMsg{err}
		}
		return tagsUpdatedMsg{}
	}
}

// moveTask moves a task to the target column
func (m Model) moveTask(task *model.Task, targetColumn int) tea.Cmd {
	newStatus := m.columns[targetColumn].Status

	return func() tea.Msg {
		err := m.db.UpdateTaskStatus(task.ID, newStatus)
		if err != nil {
			return errMsg{err}
		}
		return taskUpdatedMsg{}
	}
}
