package main

import (
	"ezshield/config"
	"ezshield/internal/app"
	"ezshield/internal/commands/sync_worker"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {

	config.InitConfigs()
	config.CreateAssetsDirectories()
	var p *tea.Program
	if len(os.Args) > 1 && os.Args[1] == "start-worker" {
		p = tea.NewProgram(sync_worker.InitialModel())
	} else {
		p = tea.NewProgram(app.InitialModel())
	}
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
