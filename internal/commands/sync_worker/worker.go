package sync_worker

import (
	"ezshield/config"
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type workerFinishedMsg struct{ err error }
type sleepFinishedMsg struct{ err error }

const sleepDuration = 20

func runWorker(m Model) tea.Cmd {
	m.err = nil

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
	cmd := exec.Command("namada", args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return workerFinishedMsg{err}
	})
}
func sleepCmd() tea.Msg {
	time.Sleep(time.Second * sleepDuration)
	return sleepFinishedMsg{}

}

type Model struct {
	sleeping bool
	err      error
}

func (m Model) Init() tea.Cmd {
	return runWorker(m)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case workerFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.sleeping = true
		return m, sleepCmd
	case sleepFinishedMsg:
		return m, runWorker(m)
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}
	if m.sleeping {
		return fmt.Sprintf("Sleeping for %d Seconds", sleepDuration)
	}
	return "Shielded Sync Command"
}

func InitialModel() Model {
	return Model{}
}
