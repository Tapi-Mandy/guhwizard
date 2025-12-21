// FILE: internal/ui/model.go
package ui

import (
	"fmt"
	"strings"

	"guhwizard/internal/config"
	"guhwizard/internal/engine"
	"guhwizard/internal/styles"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type AppState int

const (
	StateWelcome AppState = iota
	StateSelection
	StateConfirmation
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

	// UI Components
	width    int
	height   int
	list     list.Model
	progress progress.Model
	viewport viewport.Model

	// Data & Channels
	logChannel  chan string
	progChannel chan engine.ProgressMsg
	logs        []string
	showLogs    bool
	statusMsg   string
}

func NewModel(cfg *config.Config) Model {
	// 1. Setup Progress Bar
	prog := progress.New(progress.WithGradient("#f2cdcd", "#cba6f7"))

	// 2. Setup Viewport (Logs)
	vp := viewport.New(0, 0)

	// 3. Setup Channels & Engine
	logChan := make(chan string, 100)
	progChan := make(chan engine.ProgressMsg, 100)
	runner := engine.NewRunner(cfg, logChan, progChan)

	// 4. Setup List with CUSTOM DELEGATE
	l := list.New([]list.Item{}, CustomDelegate{}, 0, 0)
	l.SetShowHelp(false)
	l.Styles.Title = styles.Highlight

	m := Model{
		state:       StateWelcome,
		cfg:         cfg,
		runner:      runner,
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
	return nil
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
		currStep := m.cfg.Steps[m.currentStepIdx]

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// --- ENFORCE SELECTION FOR AUR HELPERS ---
				if currStep.ID == "aur" {
					hasSelection := false
					for _, item := range currStep.Items {
						if item.Selected {
							hasSelection = true
							m.cfg.Settings.AURHelper = item.Name
							break
						}
					}
					if !hasSelection {
						return m, nil // Don't allow skipping AUR selection
					}
				}

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

					// Single select logic
					if currStep.Type == "single" {
						// Unselect all others in BOTH the model list and the config
						for i := range currStep.Items {
							currStep.Items[i].Selected = false
						}
						// The display list model also needs its items updated
						items := m.list.Items()
						for i := range items {
							li := items[i].(listItem)
							li.configItem.Selected = false
							m.list.SetItem(i, li)
						}
					}

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
				m.state = StateInstalling
				return m, tea.Batch(
					waitForLog(m.logChannel),
					waitForProgress(m.progChannel),
					func() tea.Msg {
						err := m.runner.Install()
						return installMsg{err: err}
					},
				)
			} else if msg.String() == "esc" {
				m.state = StateSelection
				m.currentStepIdx = len(m.cfg.Steps) - 1
				m.loadCurrentStep()
				return m, nil
			}
		}

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

	// Simple heuristic: Remove trailing numbers
	displayWords := strings.Fields(displayName)
	if len(displayWords) > 1 {
		lastWord := displayWords[len(displayWords)-1]
		if lastWord == "bin" || (len(lastWord) == 1 && lastWord >= "0" && lastWord <= "9") {
			displayName = strings.Join(displayWords[:len(displayWords)-1], " ")
		}
	}

	return fmt.Sprintf("%s %s", check, displayName)
}

func (i listItem) Description() string { return i.configItem.Description }
func (i listItem) FilterValue() string { return i.configItem.Name }
