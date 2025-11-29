package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/happytaoer/cli_kanban/internal/db"
	"github.com/happytaoer/cli_kanban/internal/tui"
	"github.com/spf13/cobra"
)

var (
	dbPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cli_kanban",
		Short: "A terminal-based Kanban board",
		Long:  `cli_kanban is a beautiful TUI application for managing tasks in a Kanban board format.`,
		RunE:  runTUI,
	}

	// Get default database path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultDBPath := filepath.Join(homeDir, ".cli_kanban.db")

	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", defaultDBPath, "Path to SQLite database file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Create TUI model
	model := tui.NewModel(database)

	// Start TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
