package model

import "time"

// TaskStatus represents the status column of a task
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// Task represents a kanban task item
type Task struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Column represents a kanban column
type Column struct {
	Name   string
	Status TaskStatus
	Tasks  []Task
}

// GetAllColumns returns all three columns in order
func GetAllColumns() []Column {
	return []Column{
		{Name: "Todo", Status: StatusTodo},
		{Name: "In Progress", Status: StatusInProgress},
		{Name: "Done", Status: StatusDone},
	}
}

// NextStatus returns the next status in the workflow
func (s TaskStatus) Next() TaskStatus {
	switch s {
	case StatusTodo:
		return StatusInProgress
	case StatusInProgress:
		return StatusDone
	default:
		return StatusDone
	}
}

// PrevStatus returns the previous status in the workflow
func (s TaskStatus) Prev() TaskStatus {
	switch s {
	case StatusDone:
		return StatusInProgress
	case StatusInProgress:
		return StatusTodo
	default:
		return StatusTodo
	}
}
