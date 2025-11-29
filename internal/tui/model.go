package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/happytaoer/cli_kanban/internal/db"
	"github.com/happytaoer/cli_kanban/internal/model"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeBoard ViewMode = iota
	ViewModeAddTask
	ViewModeEditTask
	ViewModeHelp
)

// Model is the main TUI model
type Model struct {
	db            *db.DB
	columns       []model.Column
	currentColumn int
	currentTask   int
	viewMode      ViewMode
	textInput     textinput.Model
	width         int
	height        int
	err           error
}

// NewModel creates a new TUI model
func NewModel(database *db.DB) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	return Model{
		db:            database,
		columns:       model.GetAllColumns(),
		currentColumn: 0,
		currentTask:   0,
		viewMode:      ViewModeBoard,
		textInput:     ti,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.loadTasks()
}

// loadTasks loads all tasks from the database
func (m Model) loadTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.db.GetAllTasks()
		if err != nil {
			return errMsg{err}
		}
		return tasksLoadedMsg{tasks}
	}
}

// Messages
type tasksLoadedMsg struct {
	tasks []model.Task
}

type taskCreatedMsg struct {
	task *model.Task
}

type taskUpdatedMsg struct{}

type taskDeletedMsg struct{}

type errMsg struct {
	err error
}

// getCurrentTask returns the currently selected task
func (m *Model) getCurrentTask() *model.Task {
	col := &m.columns[m.currentColumn]
	if len(col.Tasks) == 0 || m.currentTask >= len(col.Tasks) {
		return nil
	}
	return &col.Tasks[m.currentTask]
}

// organizeTasks organizes tasks into columns by status
func (m *Model) organizeTasks(tasks []model.Task) {
	// Reset all columns
	for i := range m.columns {
		m.columns[i].Tasks = []model.Task{}
	}

	// Organize tasks by status
	for _, task := range tasks {
		for i := range m.columns {
			if m.columns[i].Status == task.Status {
				m.columns[i].Tasks = append(m.columns[i].Tasks, task)
				break
			}
		}
	}

	// Ensure currentTask is within bounds
	if m.currentTask >= len(m.columns[m.currentColumn].Tasks) {
		m.currentTask = len(m.columns[m.currentColumn].Tasks) - 1
	}
	if m.currentTask < 0 {
		m.currentTask = 0
	}
}
