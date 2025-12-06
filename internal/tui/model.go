package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	ViewModeEditDescription
	ViewModeEditTags
	ViewModeConfirmDelete
	ViewModeHelp
	ViewModeSearch
)

// Model is the main TUI model
type Model struct {
	db              *db.DB
	columns         []model.Column
	currentColumn   int
	currentTask     int
	scrollOffsets   []int // scroll offset per column
	viewMode        ViewMode
	currentTime     time.Time
	pendingDeleteID int64 // task ID pending deletion confirmation
	followTaskID    int64 // task ID to follow after reload
	textInput       textinput.Model
	textArea        textarea.Model
	searchInput     textinput.Model
	searchQuery     string // active search filter
	viewport        viewport.Model
	width           int
	height          int
	ready           bool // viewport ready flag
	err             error
}

// clockTickCmd creates a command that emits time ticks every second
func clockTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockTickMsg(t)
	})
}

// NewModel creates a new TUI model
func NewModel(database *db.DB) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	ta := textarea.New()
	ta.Placeholder = "Enter task description..."
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.CharLimit = 2000

	si := textinput.New()
	si.Placeholder = "Search tasks..."
	si.CharLimit = 100
	si.Width = 30

	return Model{
		db:            database,
		columns:       model.GetAllColumns(),
		currentColumn: 0,
		currentTask:   0,
		scrollOffsets: make([]int, 3), // one per column
		currentTime:   time.Now(),
		viewMode:      ViewModeBoard,
		textInput:     ti,
		textArea:      ta,
		searchInput:   si,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadTasks(), clockTickCmd())
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

type descriptionUpdatedMsg struct{}

type tagsUpdatedMsg struct{}

type clockTickMsg time.Time

type errMsg struct {
	err error
}

// maxVisibleTasks is the maximum number of tasks visible per column
const maxVisibleTasks = 10

// getCurrentTask returns the currently selected task
func (m *Model) getCurrentTask() *model.Task {
	col := &m.columns[m.currentColumn]
	if len(col.Tasks) == 0 || m.currentTask >= len(col.Tasks) {
		return nil
	}
	return &col.Tasks[m.currentTask]
}

// ensureTaskVisible adjusts scroll offset to keep current task visible
func (m *Model) ensureTaskVisible() {
	offset := m.scrollOffsets[m.currentColumn]

	// If current task is above visible area, scroll up
	if m.currentTask < offset {
		m.scrollOffsets[m.currentColumn] = m.currentTask
	}

	// If current task is below visible area, scroll down
	if m.currentTask >= offset+maxVisibleTasks {
		m.scrollOffsets[m.currentColumn] = m.currentTask - maxVisibleTasks + 1
	}
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

	// If we're following a task after move, find its position
	if m.followTaskID != 0 {
		found := false
		for i, task := range m.columns[m.currentColumn].Tasks {
			if task.ID == m.followTaskID {
				m.currentTask = i
				found = true
				break
			}
		}
		m.followTaskID = 0 // Clear after finding
		if found {
			m.ensureTaskVisible()
			return
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
