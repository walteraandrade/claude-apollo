package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/walter/apollo/internal/config"
	"github.com/walter/apollo/internal/db"
	"github.com/walter/apollo/internal/notifier"
	"github.com/walter/apollo/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		cfg.RepoPath = os.Args[1]
	}

	if cfg.RepoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
			os.Exit(1)
		}
		cfg.RepoPath = cwd
	}

	if err := os.MkdirAll(config.ApolloDir(), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	database, err := db.Open(config.DBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	n := notifier.New()
	model := tui.NewModel(cfg, database, n)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "apollo: %v\n", err)
		os.Exit(1)
	}
}
