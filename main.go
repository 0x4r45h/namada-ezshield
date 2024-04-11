package main

import (
	"bufio"
	"ezshield/config"
	"ezshield/internal/app"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
)

func main() {

	config.InitConfigs()
	config.CreateAssetsDirectories()
	var p *tea.Program
	if len(os.Args) > 1 && os.Args[1] == "start-worker" {
		// TODO: this should work, but crash with nil pointer error,ask in their github
		//p = tea.NewProgram(sync_worker.InitialModel(), tea.WithoutRenderer(), tea.WithInput(nil))
		// naive workaround until find a solution for above
		args := []string{
			"client",
			"shielded-sync",
			"--node",
			config.Cfg.Namada.Node,
			"--chain-id",
			config.Cfg.Namada.ChainId,
		}
		if config.Cfg.Namada.ChainId == "shielded-expedition.88f17d1d14" {
			args = append(args, []string{"--from-height", "237907"}...)
		}
		for {
			cmd := exec.Command("namada", args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Println("Error creating pipe:", err)
				return
			}
			if err := cmd.Start(); err != nil {
				fmt.Println("Error starting command:", err)
				return
			}
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading from pipe:", err)
			}

			if err := cmd.Wait(); err != nil {
				fmt.Println("Error waiting for command to finish:", err)
			}
		}
	} else {
		p = tea.NewProgram(app.InitialModel())
	}
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
