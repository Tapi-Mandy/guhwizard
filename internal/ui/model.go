// FILE: internal/ui/model.go
package ui

import (
	"fmt"
	"strings"

	"guhwizard/internal/config"
	"guhwizard/internal/engine"
	"guhwizard/internal/installer"
	"guhwizard/internal/styles"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type AppState int

const (
	StateWelcome AppState = iota
	StateSelection
	StateConfirmation
	StatePassword
	StateInstalling
	StateDone
)

type installMsg struct{ err error }
type logMsg string

type Model struct {
	state          AppState
	cfg            *config.Config
	currentStepIdx int
	runner         *engine.Runner
	password       string

	// UI Components
	width     int
	height    int
	textInput textinput.Model
	list      list.Model
	progress  progress.Model
	viewport  viewport.Model

	// Data & Channels
	logChannel  chan string
	progChannel chan engine.ProgressMsg
	logs        []string
	showLogs    bool
	statusMsg   string
}

func NewModel(cfg *config.Config) Model {
	// 1. Setup Text Input (Password)
	ti := textinput.New()
	ti.Placeholder = "Sudo Password"
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 156
	ti.Width = 30

	// 2. Setup Progress Bar
	prog := progress.New(progress.WithGradient("#f2cdcd", "#cba6f7"))

	// 3. Setup Viewport (Logs)
	vp := viewport.New(0, 0)

	// 4. Setup Channels & Engine
	logChan := make(chan string, 100)
	progChan := make(chan engine.ProgressMsg, 100)
	runner := engine.NewRunner(cfg, logChan, progChan)

	// 5. Setup List with CUSTOM DELEGATE
	// Ensure you created internal/ui/delegate.go for this to work!
	l := list.New([]list.Item{}, CustomDelegate{}, 0, 0)
	l.SetShowHelp(false)
	l.Styles.Title = styles.Highlight

	m := Model{
		state:       StateWelcome,
		cfg:         cfg,
		runner:      runner,
		textInput:   ti,
		progress:    prog,
		viewport:    vp,
		list:        l,
		logChannel:  logChan,
		progChannel: progChan,
	}

	return m
}

// -- Commands --

func waitForLog(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-sub)
	}
}

func waitForProgress(sub chan engine.ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Resize sub-components
		m.list.SetWidth(msg.Width - 10)
		m.list.SetHeight(14)
		m.viewport.Width = msg.Width - 10
		m.viewport.Height = 10
		m.progress.Width = msg.Width - 20

	// --- Channel Handling ---
	case logMsg:
		m.logs = append(m.logs, string(msg))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, waitForLog(m.logChannel)

	case engine.ProgressMsg:
		cmd = m.progress.SetPercent(msg.CurrentPercent)
		cmds = append(cmds, cmd)
		m.statusMsg = msg.CurrentStep
		cmds = append(cmds, waitForProgress(m.progChannel))
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case installMsg:
		if msg.err != nil {
			m.logs = append(m.logs, styles.Error.Render(fmt.Sprintf("\nERROR: %v", msg.err)))
			m.showLogs = true
		} else {
			m.state = StateDone
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// --- State Machine ---
	switch m.state {
	case StateWelcome:
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			m.state = StateSelection
			m.currentStepIdx = 0
			m.loadCurrentStep()
			return m, nil
		}

	case StateSelection:
		if m.currentStepIdx >= len(m.cfg.Steps) {
			m.state = StateConfirmation
			return m, nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.currentStepIdx++
				if m.currentStepIdx < len(m.cfg.Steps) {
					m.loadCurrentStep()
				} else {
					m.state = StateConfirmation
					return m, nil
				}
			case " ":
				if len(m.list.Items()) > 0 {
					idx := m.list.Index()
					itm := m.list.SelectedItem().(listItem)
					itm.configItem.Selected = !itm.configItem.Selected
					m.list.SetItem(idx, itm)
				}
			}
		}
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case StateConfirmation:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "enter" {
				m.state = StatePassword
				m.textInput.Focus()
				return m, nil
			} else if msg.String() == "esc" {
				// Backtrack could be added here
			}
		}

	case StatePassword:
		m.textInput, cmd = m.textInput.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			m.password = m.textInput.Value()

			// --- FIXED VALIDATION LOGIC ---
			// We use StartSudoKeepAlive to check if the password is correct.
			err := installer.CurrentSession.StartSudoKeepAlive(m.password)

			if err != nil {
				// Password Invalid
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Incorrect. Try again."
			} else {
				// Password Valid!
				// Stop the keepalive immediately, because runner.Install()
				// will start its own fresh session.
				installer.CurrentSession.StopSudo()

				m.state = StateInstalling
				return m, tea.Batch(
					waitForLog(m.logChannel),
					waitForProgress(m.progChannel),
					func() tea.Msg {
						err := m.runner.Install(m.password)
						return installMsg{err: err}
					},
				)
			}
		}
		return m, cmd

	case StateInstalling:
		if msg, ok := msg.(tea.KeyMsg); ok && (msg.String() == "v" || msg.String() == "V") {
			m.showLogs = !m.showLogs
		}
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case StateDone:
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) loadCurrentStep() {
	step := m.cfg.Steps[m.currentStepIdx]
	items := []list.Item{}
	for i := range step.Items {
		items = append(items, listItem{configItem: &step.Items[i]})
	}

	cmd := m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("Step %d/%d: %s", m.currentStepIdx+1, len(m.cfg.Steps), step.Title)
	m.list.ResetSelected()
	_ = cmd
}

type listItem struct {
	configItem *config.Item
}

func (i listItem) Title() string {
	check := "[ ]"
	if i.configItem.Selected {
		check = "[x]"
	}

	displayName := i.configItem.Name
	// Prettify: replace dashes with spaces
	displayName = strings.ReplaceAll(displayName, "-", " ")

	// Basic cleanup: remove common version suffixes manually or via logic
	// e.g. "sublime text 4" -> "sublime text" (if desired strictly)
	// For now, replacing dashes is the main request + removing strict version numbers if easy.
	// Let's trim trailing digits if separated by space?
	// User example: "sublime-text-4" -> "sublime text".

	// Simple heuristic: Remove trailing numbers
	displayWords := strings.Fields(displayName)
	if len(displayWords) > 1 {
		lastWord := displayWords[len(displayWords)-1]
		// If last word is a single digit or "bin", remove it?
		// "bin" is common in AUR.
		if lastWord == "bin" || (len(lastWord) == 1 && lastWord >= "0" && lastWord <= "9") {
			displayName = strings.Join(displayWords[:len(displayWords)-1], " ")
		}
	}

	return fmt.Sprintf("%s %s", check, displayName)
}

func (i listItem) Description() string { return i.configItem.Description }
func (i listItem) FilterValue() string { return i.configItem.Name }
