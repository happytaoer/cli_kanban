# cli_kanban

A terminal-based Kanban board management tool built with Go, featuring a beautiful TUI interface.

## Features

- 📋 **Three-column board**: Todo / In Progress / Done
- ✨ **Full CRUD operations**: Add, edit, and delete tasks
- 🎨 **Beautiful TUI interface**: Built with Bubble Tea framework
- 💾 **SQLite persistence**: Data automatically saved to local database
- ⌨️ **Keyboard shortcuts**: Efficient keyboard navigation

## Installation

### Prerequisites

- Go 1.21 or higher
- GCC (for compiling SQLite)

### Build

```bash
# Clone or navigate to project directory
cd cli_kanban

# Download dependencies
go mod tidy

# Build
go build -o cli_kanban

# Run
./cli_kanban
```

## Usage

### Launch Application

```bash
# Use default database path (~/.cli_kanban.db)
./cli_kanban

# Specify custom database path
./cli_kanban --db /path/to/kanban.db
```

### Keyboard Shortcuts

#### Navigation
- `←` / `→` or `h` / `l` - Switch between columns
- `↑` / `↓` or `j` / `k` - Move between tasks

#### Actions
- `a` - Add new task to current column
- `e` or `Enter` - Edit selected task
- `d` or `Delete` - Delete selected task
- `m` - Move task to next column

#### Other
- `?` - Show help
- `q` or `Ctrl+C` - Quit application
- `Esc` - Cancel current action or quit

## Project Structure

```
cli_kanban/
├── main.go              # Entry point and Cobra commands
├── go.mod               # Go module dependencies
├── internal/
│   ├── db/
│   │   └── sqlite.go    # SQLite database operations
│   ├── model/
│   │   └── task.go      # Data model definitions
│   └── tui/
│       ├── model.go     # Bubble Tea model
│       ├── update.go    # Event handling logic
│       └── view.go      # View rendering
└── README.md
```

## Tech Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Styling and layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components
- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[SQLite](https://github.com/mattn/go-sqlite3)** - Data persistence

## Data Model

### Task

| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER | Auto-increment primary key |
| title | TEXT | Task title |
| status | TEXT | Task status (todo/in_progress/done) |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update timestamp |

## Development

```bash
# Run (development mode)
go run main.go

# Format code
go fmt ./...

# Run tests
go test ./...
```

## License

MIT
