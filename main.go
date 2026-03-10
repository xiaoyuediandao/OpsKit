package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"opskit/internal/ai"
	"opskit/internal/config"
	"opskit/internal/state"
	"opskit/internal/tui"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	startInSetup := cfg == nil || cfg.APIKey == ""

	// Load persisted state
	st, _ := state.Load()
	if st == nil {
		st = state.DefaultState()
	}

	lobster := tui.NewLobster(st)

	// Create AI client — always apply defaults for missing fields
	var aiClient *ai.Client
	if cfg != nil && cfg.APIKey != "" {
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = config.DefaultBaseURL
		}
		model := cfg.Model
		if model == "" {
			model = config.DefaultModel
		}
		aiClient = ai.NewClient(cfg.APIKey, baseURL, model)
	} else {
		aiClient = ai.NewClient("", config.DefaultBaseURL, config.DefaultModel)
	}

	// Create model
	m := tui.InitialModel(aiClient, lobster, st, startInSetup)

	// Create program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running OpsKit: %v\n", err)
		os.Exit(1)
	}
}
