// FILE: cmd/guhwizard/main.go
package main

import (
	"flag"
	"fmt"
	"os"

	"guhwizard/internal/config"
	"guhwizard/internal/root"
	"guhwizard/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	rootSetup := flag.Bool("root-setup", false, "Run root setup for sudo persistence")
	flag.Parse()

	if *rootSetup {
		if err := root.ConfigureSudoTimestamp(); err != nil {
			fmt.Fprintf(os.Stderr, "Root setup failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 1. Load the Installation Blueprint
	// In a real release, you might embed this file into the binary using `//go:embed`
	// so you don't need the external file at runtime.
	cfg, err := config.Load("install_config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Println("Make sure 'install_config.yaml' is in the current directory.")
		os.Exit(1)
	}

	// 2. Initialize the UI Model with the Config
	model := ui.NewModel(cfg)

	// 3. Run the Bubble Tea Program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
