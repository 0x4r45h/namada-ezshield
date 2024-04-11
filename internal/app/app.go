package app

import (
	"ezshield/internal/commands/sibc_from_namada"
	"ezshield/internal/commands/sibc_to_namada"
	"ezshield/internal/commands/sync_worker"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var choices = []string{"Shielded IBC - Osmosis -> Namada", "Shielded IBC - Namada -> Osmosis", "Run Sync Shield Worker"}

type Model struct {
	cursor int
	choice string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			m.choice = choices[m.cursor]
			switch m.choice {
			case "Shielded IBC - Osmosis -> Namada":
				return sibc_to_namada.InitialModel(), tea.ClearScreen
			case "Shielded IBC - Namada -> Osmosis":
				return sibc_from_namada.InitialModel(), tea.ClearScreen
			case "Run Sync Shield Worker":
				m := sync_worker.InitialModel()
				return m, m.Init()
			}

			return m, nil

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString("Select one of the options\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("[*] ")
		} else {
			s.WriteString("[ ] ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}

func InitialModel() Model {
	return Model{}
}
