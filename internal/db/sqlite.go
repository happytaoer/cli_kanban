package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/happytaoer/cli_kanban/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initTables(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initTables creates the necessary tables if they don't exist
func (db *DB) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// CreateTask creates a new task
func (db *DB) CreateTask(title string, status model.TaskStatus) (*model.Task, error) {
	now := time.Now()
	result, err := db.conn.Exec(
		"INSERT INTO tasks (title, status, created_at, updated_at) VALUES (?, ?, ?, ?)",
		title, status, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &model.Task{
		ID:        id,
		Title:     title,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetAllTasks retrieves all tasks
func (db *DB) GetAllTasks() ([]model.Task, error) {
	rows, err := db.conn.Query(
		"SELECT id, title, status, created_at, updated_at FROM tasks ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByStatus retrieves tasks by status
func (db *DB) GetTasksByStatus(status model.TaskStatus) ([]model.Task, error) {
	rows, err := db.conn.Query(
		"SELECT id, title, status, created_at, updated_at FROM tasks WHERE status = ? ORDER BY created_at DESC",
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates a task
func (db *DB) UpdateTask(id int64, title string, status model.TaskStatus) error {
	result, err := db.conn.Exec(
		"UPDATE tasks SET title = ?, status = ?, updated_at = ? WHERE id = ?",
		title, status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// UpdateTaskStatus updates only the status of a task
func (db *DB) UpdateTaskStatus(id int64, status model.TaskStatus) error {
	result, err := db.conn.Exec(
		"UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask deletes a task
func (db *DB) DeleteTask(id int64) error {
	result, err := db.conn.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
