package sibc_from_namada

import (
	"bytes"
	"ezshield/config"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os/exec"
	"strings"
)

type (
	errMsg error
)

type cmdErrMsg struct {
	Err    error
	Stderr bytes.Buffer
}

func (r cmdErrMsg) Error() string {
	return fmt.Sprintf("%s\n\n%s\n\n", r.Err, r.Stderr.String())
}

const (
	sourceAccount = iota
	receiverAddr  = iota
	token
	amount
	namadaFlags
)

var (
	header = lipgloss.NewStyle().Foreground(config.White)

	inputStyle    = lipgloss.NewStyle().Foreground(config.NamadaYellow)
	continueStyle = lipgloss.NewStyle().Foreground(config.DarkGray)
)

type Model struct {
	spinner    spinner.Model
	inputs     []textinput.Model
	focused    int
	err        error
	txResult   *TxResult
	inProgress bool
}

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(config.NamadaYellow)
	var inputs = make([]textinput.Model, 5)

	inputs[sourceAccount] = textinput.New()
	inputs[sourceAccount].Focus()
	inputs[sourceAccount].Placeholder = "The source account name you want to send funds from"
	inputs[sourceAccount].CharLimit = 100
	inputs[sourceAccount].Width = 100
	inputs[sourceAccount].Prompt = ""

	inputs[receiverAddr] = textinput.New()
	inputs[receiverAddr].Placeholder = "The receiverAddr address on the destination chain as string."
	inputs[receiverAddr].CharLimit = 140
	inputs[receiverAddr].Width = 140
	inputs[receiverAddr].Prompt = ""

	inputs[token] = textinput.New()
	inputs[token].Placeholder = "The transfer token"
	inputs[token].CharLimit = 68
	inputs[token].Width = 76
	inputs[token].Prompt = ""

	inputs[amount] = textinput.New()
	inputs[amount].Placeholder = "The amount to transfer in decimal"
	inputs[amount].CharLimit = 20
	inputs[amount].Width = 20
	inputs[amount].Prompt = ""

	inputs[namadaFlags] = textinput.New()
	inputs[namadaFlags].Placeholder = "Append any custom namadaFlags as you do in cli"
	inputs[namadaFlags].CharLimit = 250
	inputs[namadaFlags].Width = 250
	inputs[namadaFlags].Prompt = ""

	return Model{
		spinner: s,
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m Model) Init() tea.Cmd {
	//return tea.Batch(textinput.Blink, m.spinner.Tick)
	return textinput.Blink
	//return m.spinner.Tick
}

type CmdResult struct {
	Stdout, Stderr bytes.Buffer
}
type NamadaAddr struct {
	Alias, Addr string
}
type NamadaMemo struct {
	Memo string
}
type TxResult struct {
	Result string
}

func sendNamadaShieldedIBCCMD(m Model) tea.Cmd {
	return func() tea.Msg {
		txRes, cmdr, err := sendNamadaShieldedTx(m)
		if err != nil {
			return cmdErrMsg{Err: err, Stderr: cmdr.Stderr}
		}
		return TxResultMsg{txRes}
	}
}

type NamadaCreateShieldedMemoMsg struct {
	NamadaMemo
}
type TxResultMsg struct {
	TxResult
}

func sendNamadaShieldedTx(m Model) (TxResult, CmdResult, error) {
	cmdFlags := strings.Split(m.inputs[namadaFlags].Value(), " ")
	args := []string{
		"client",
		"ibc-transfer",
		"--node",
		config.Cfg.Namada.Node,
		"--chain-id",
		config.Cfg.Namada.ChainId,
		"--amount",
		m.inputs[amount].Value(),
		"--source",
		m.inputs[sourceAccount].Value(),
		"--receiver",
		m.inputs[receiverAddr].Value(),
		"--token",
		m.inputs[token].Value(),
		"--channel-id",
		config.Cfg.Namada.ChannelOsmosis.ChannelId,
		"--disposable-gas-payer",
		"--gas-spending-key",
		m.inputs[sourceAccount].Value(),
	}
	if cmdFlags[0] != "" {
		args = append(args, cmdFlags...)
	}
	cmd := exec.Command("namada", args...)
	txRes := TxResult{}
	cmdr := CmdResult{}
	cmd.Stdout = &cmdr.Stdout
	cmd.Stderr = &cmdr.Stderr
	err := cmd.Run()
	if err != nil {
		return txRes, cmdr, err
	}
	txRes.Result = cmdr.Stdout.String()
	return txRes, cmdr, err

}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if !m.inProgress && (m.focused == len(m.inputs)-1) && m.txResult == nil {
				m.inProgress = true
				return m, tea.Batch(sendNamadaShieldedIBCCMD(m), m.spinner.Tick)
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	case TxResultMsg:
		m.inProgress = false
		m.txResult = &msg.TxResult
		return m, nil
	case errMsg:
		m.err = msg
		m.inProgress = false
		return m, tea.Quit
	default:
		if m.inProgress {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := strings.Builder{}
	if m.inProgress {
		s.WriteString(fmt.Sprintf("\n\n   %s Task is in progress\n\n", m.spinner.View()))
		return s.String()
	}
	if m.err != nil {
		s.WriteString(fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err))
		return s.String()
	}
	if m.txResult != nil {
		r := fmt.Sprintf("Tx result is :\n\n`%s`\n\n", m.txResult.Result)
		s.WriteString(r)
	}
	if s.Len() > 0 {
		return s.String()
	}
	return fmt.Sprintf(
		`
 %s

 %s  
 %s

 %s
 %s

 %s  %s
 %s  %s

 %s
 %s

 %s
`,
		header.Bold(true).Width(150).Render("Enter required inputs, any custom flags can be passed to the 'flags' inputs:"),
		inputStyle.Width(100).Render("Source Account"),
		m.inputs[sourceAccount].View(),
		inputStyle.Width(140).Render("Receiver Address"),
		m.inputs[receiverAddr].View(),
		inputStyle.Width(77).Render("Token"),
		inputStyle.Width(20).Render("Amount"),
		m.inputs[token].View(),
		m.inputs[amount].View(),
		inputStyle.Width(40).Render("Custom Flags (Namada)"),
		m.inputs[namadaFlags].View(),
		continueStyle.Render("Press Enter to submit the tx"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *Model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *Model) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
