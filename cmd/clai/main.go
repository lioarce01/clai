package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lioarce01/clai/internal/config"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/storage"
	"github.com/lioarce01/clai/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize storage
	store, err := storage.NewJSONStore(cfg.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	// Create LLM client (validation of API key happens lazily on first request)
	llmClient := llm.NewClient(cfg.API.APIKey, cfg.API.BaseURL)

	// Build TUI app
	app, err := tui.New(cfg, llmClient, store)
	if err != nil {
		return fmt.Errorf("init tui: %w", err)
	}

	// Run the Bubble Tea program
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run program: %w", err)
	}

	return nil
}
