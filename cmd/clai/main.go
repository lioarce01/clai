package main

import (
	"flag"
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
	var (
		flagAPIKey  = flag.String("api-key", "", "API key (overrides config and OPENAI_API_KEY)")
		flagBaseURL = flag.String("base-url", "", "API base URL (overrides config and OPENAI_BASE_URL)")
		flagModel   = flag.String("model", "", "Model name (overrides config)")
	)
	flag.Parse()

	// Load base configuration (file + env vars)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// CLI flags take highest priority
	if *flagAPIKey != "" {
		cfg.API.APIKey = *flagAPIKey
	}
	if *flagBaseURL != "" {
		cfg.API.BaseURL = *flagBaseURL
	}
	if *flagModel != "" {
		cfg.Model.Name = *flagModel
	}

	// Initialize storage
	store, err := storage.NewJSONStore(cfg.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	// Create LLM client
	llmClient := llm.NewClient(cfg.API.APIKey, cfg.API.BaseURL)

	// Build and run TUI
	app, err := tui.New(cfg, llmClient, store)
	if err != nil {
		return fmt.Errorf("init tui: %w", err)
	}

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
