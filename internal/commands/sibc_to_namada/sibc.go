package sibc_to_namada

import (
	"bytes"
	"ezshield/config"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"os"
	"os/exec"
	"regexp"
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
	receiverAccount
	token
	amount
	namadaGenAddressFlags
	namadaFlags
	osmosisFlags
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
	namadaAddr *NamadaAddr
	namadaMemo *NamadaMemo
	txResult   *TxResult
	inProgress bool
}

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(config.NamadaYellow)
	var inputs = make([]textinput.Model, 7)

	inputs[sourceAccount] = textinput.New()
	inputs[sourceAccount].Focus()
	inputs[sourceAccount].Placeholder = "The source account name you want to send funds from"
	inputs[sourceAccount].CharLimit = 100
	inputs[sourceAccount].Width = 100
	inputs[sourceAccount].Prompt = ""

	inputs[receiverAccount] = textinput.New()
	inputs[receiverAccount].Placeholder = "The receiver address on the destination chain as string."
	inputs[receiverAccount].CharLimit = 140
	inputs[receiverAccount].Width = 140
	inputs[receiverAccount].Prompt = ""

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

	inputs[namadaGenAddressFlags] = textinput.New()
	inputs[namadaGenAddressFlags].Placeholder = "Append any custom Flags for `namada gen-payment-addr` command"
	inputs[namadaGenAddressFlags].CharLimit = 500
	inputs[namadaGenAddressFlags].Width = 500
	inputs[namadaGenAddressFlags].Prompt = ""

	inputs[namadaFlags] = textinput.New()
	inputs[namadaFlags].Placeholder = "Append any custom Flags for `namada` command"
	inputs[namadaFlags].CharLimit = 500
	inputs[namadaFlags].Width = 500
	inputs[namadaFlags].Prompt = ""

	inputs[osmosisFlags] = textinput.New()
	inputs[osmosisFlags].Placeholder = "Append any custom Flags for `osmosis` command"
	inputs[osmosisFlags].CharLimit = 250
	inputs[osmosisFlags].Width = 250
	inputs[osmosisFlags].Prompt = ""

	return Model{
		spinner: s,
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
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

func createShieldedAddrCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		namadaAddr, cmdr, err := GenShieldedAddr(m)
		if err != nil {
			return cmdErrMsg{Err: err, Stderr: cmdr.Stderr}
		}
		return NamadaCreateShieldedAddrMsg{namadaAddr}
	}
}

type NamadaCreateShieldedAddrMsg struct {
	NamadaAddr
}

func createShieldedMemo(m Model) tea.Cmd {
	return func() tea.Msg {
		nm, cmdr, err := GenShieldedMemo(m)
		if err != nil {
			return cmdErrMsg{Err: err, Stderr: cmdr.Stderr}
		}
		return NamadaCreateShieldedMemoMsg{nm}
	}
}
func sendOsmosisShieldedIBC(m Model) tea.Cmd {
	return func() tea.Msg {
		txRes, cmdr, err := sendOsmosisShieldedTx(m)
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

func GenShieldedAddr(m Model) (NamadaAddr, CmdResult, error) {
	namaddr := NamadaAddr{}
	accName := m.inputs[receiverAccount].Value()
	salias := fmt.Sprintf("%s-%s", accName, uuid.New().String())
	cmdFlags := strings.Split(m.inputs[namadaGenAddressFlags].Value(), " ")
	cmdArgs := []string{
		"wallet", "gen-payment-addr", "--key", accName, "--alias", salias, "--chain-id",
		config.Cfg.Namada.ChainId,
	}
	if cmdFlags[0] != "" {
		cmdArgs = append(cmdArgs, cmdFlags...)
	}
	cmd := exec.Command("namada", cmdArgs...)
	cmdr := CmdResult{}
	cmd.Stdout = &cmdr.Stdout
	cmd.Stderr = &cmdr.Stderr
	err := cmd.Run()
	if err != nil {
		return namaddr, cmdr, err
	}
	re := regexp.MustCompile(`znam\w{78}`)
	address := re.FindString(cmdr.Stdout.String())
	namaddr.Addr = address
	namaddr.Alias = salias
	return namaddr, cmdr, err
}

func GenShieldedMemo(m Model) (NamadaMemo, CmdResult, error) {
	cmdFlags := strings.Split(m.inputs[namadaFlags].Value(), " ")
	cmdArgs := []string{
		"client",
		"ibc-gen-shielded",
		"--node",
		config.Cfg.Namada.Node,
		"--output-folder-path",
		config.GetTempDirPath(),
		"--target",
		m.namadaAddr.Addr,
		"--token",
		m.inputs[token].Value(),
		"--amount",
		m.inputs[amount].Value(),
		"--port-id",
		"transfer",
		"--channel-id",
		config.Cfg.Namada.ChannelOsmosis.ChannelId,
	}
	if cmdFlags[0] != "" {
		cmdArgs = append(cmdArgs, cmdFlags...)
	}
	cmd := exec.Command("namada", cmdArgs...)
	nm := NamadaMemo{}
	cmdr := CmdResult{}
	cmd.Stdout = &cmdr.Stdout
	cmd.Stderr = &cmdr.Stderr
	err := cmd.Run()
	if err != nil {
		return nm, cmdr, err
	}
	re := regexp.MustCompile(`/.*?\.memo`)
	p := re.FindString(cmdr.Stdout.String())
	content, err := os.ReadFile(p)
	if err != nil {
		return nm, cmdr, err
	}
	nm.Memo = string(content)
	return nm, cmdr, err

}

func sendOsmosisShieldedTx(m Model) (TxResult, CmdResult, error) {
	cmdFlags := strings.Split(m.inputs[osmosisFlags].Value(), " ")
	cmdArgs := []string{
		"tx",
		"ibc-transfer",
		"transfer",
		"--node",
		config.Cfg.Osmosis.Node,
		"--chain-id",
		config.Cfg.Osmosis.ChainId,
		"--memo",
		m.namadaMemo.Memo,
		"--keyring-backend",
		"test",
		"--from",
		m.inputs[sourceAccount].Value(),
		"--fees",
		"500uosmo",
		"transfer",
		config.Cfg.Osmosis.ChannelNamada.ChannelId,
		m.namadaAddr.Addr,
		fmt.Sprintf("%s%s", m.inputs[amount].Value(), m.inputs[token].Value()),
		"-y",
	}
	if cmdFlags[0] != "" {
		cmdArgs = append(cmdArgs, cmdFlags...)
	}
	cmd := exec.Command("osmosisd", cmdArgs...)
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
			if !m.inProgress && (m.focused == len(m.inputs)-1) && m.namadaAddr == nil {
				m.inProgress = true
				return m, tea.Batch(createShieldedAddrCmd(m), m.spinner.Tick)
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

	case NamadaCreateShieldedAddrMsg:
		m.namadaAddr = &msg.NamadaAddr
		return m, createShieldedMemo(m)
	case NamadaCreateShieldedMemoMsg:
		m.namadaMemo = &msg.NamadaMemo
		return m, sendOsmosisShieldedIBC(m)
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
	if m.namadaAddr != nil {
		s.WriteString(fmt.Sprintf("A new namada address is generated.\n\nAlias is: `%s`\n\nAddress is: `%s`\n\n", m.namadaAddr.Alias, m.namadaAddr.Addr))
	}
	if m.namadaMemo != nil {
		s.WriteString(fmt.Sprintf("The Memo for tx is :\n\n`%s...%s`\n\n", m.namadaMemo.Memo[:20], m.namadaMemo.Memo[len(m.namadaMemo.Memo)-20:]))
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
 %s

 %s
 %s

 %s
`,
		header.Bold(true).Width(150).Render("Enter required inputs, any custom flags can be passed to the 'flags' inputs:"),
		inputStyle.Width(100).Render("Source Account"),
		m.inputs[sourceAccount].View(),
		inputStyle.Width(140).Render("Receiver Account (Shielded Account NAME)"),
		m.inputs[receiverAccount].View(),
		inputStyle.Width(77).Render("Token"),
		inputStyle.Width(20).Render("Amount"),
		m.inputs[token].View(),
		m.inputs[amount].View(),
		inputStyle.Width(80).Render("Custom Flags (Namada gen-payment-addr command)"),
		m.inputs[namadaGenAddressFlags].View(),
		inputStyle.Width(80).Render("Custom Flags (Namada)"),
		m.inputs[namadaFlags].View(),
		inputStyle.Width(80).Render("Custom Flags (Osmosis)"),
		m.inputs[osmosisFlags].View(),
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
