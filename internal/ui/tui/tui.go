// Package tui owns the terminal UI frontend. It is a dumb view: it calls engine
// methods and renders text. It does NOT contain game logic.
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/engine/sovereign"
	"github.com/mirageglobe/teio-senki/internal/models"
)

type screen int

const (
	screenSplash screen = iota
	screenMenu
	screenScenario
	screenSovereign
	screenBriefing
	screenGameA // cycle A: world update summary
	screenGameB // cycle B: player commands
	screenGameC // cycle C: settlement results
)

type model struct {
	screen      screen
	ledger      *ledger.Ledger
	engine      *sovereign.Engine
	cursor      int
	scrollOffset int
	scenarioIdx int
	lords       []models.Officer
	chosenLord  string
	cycleALogs  []models.LogEntry
	cycleCLogs  []models.LogEntry
	input       string
	feedback    string
	width       int
	height      int
}

// Run initialises the Bubble Tea program and blocks until the player exits.
func Run(l *ledger.Ledger) error {
	m := model{ledger: l, lords: l.Lords()}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenSplash:
		if key == "enter" || key == " " {
			m.screen, m.cursor = screenMenu, 0
		}
	case screenMenu:
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				m.screen, m.cursor = screenScenario, 0
			} else {
				return m, tea.Quit
			}
		}
	case screenScenario:
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(allScenarios)-1 {
				m.cursor++
			}
		case "enter":
			if allScenarios[m.cursor].unlocked {
				m.scenarioIdx = m.cursor
				m.ledger.Year = allScenarios[m.cursor].year
				m.screen, m.cursor = screenSovereign, 0
			}
		}
	case screenSovereign:
		visibleHeight := m.height - 7
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.lords)-1 {
				m.cursor++
				if m.cursor >= m.scrollOffset+visibleHeight {
					m.scrollOffset++
				}
			}
		case "enter":
			if len(m.lords) > 0 {
				m.chosenLord = m.lords[m.cursor].Name
				m.engine = sovereign.New(m.ledger, m.chosenLord)
				m.screen = screenBriefing
			}
		}
	case screenBriefing:
		if key == "enter" || key == " " {
			m.engine.StartTurn()
			m.cycleALogs = tailLogs(m.engine.GetState().Logs, 8)
			m.screen = screenGameA
		}
	case screenGameA:
		if key == "enter" || key == " " {
			m.feedback = ""
			m.screen = screenGameB
		}
	case screenGameB:
		switch key {
		case "enter":
			m = m.execCommand()
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(key) == 1 {
				m.input += key
			}
		}
	case screenGameC:
		if key == "enter" || key == " " {
			m.engine.StartTurn()
			m.cycleALogs = tailLogs(m.engine.GetState().Logs, 8)
			m.screen = screenGameA
		}
	}
	return m, nil
}

func (m model) execCommand() model {
	parts := strings.Fields(m.input)
	m.input = ""
	if len(parts) == 0 {
		return m
	}
	switch parts[0] {
	case "end":
		m.cycleCLogs = m.engine.SettleTurn()
		m.screen = screenGameC
	case "build":
		if len(parts) < 3 {
			m.feedback = "usage: build <city> ag|com|def"
			return m
		}
		typeMap := map[string]string{"ag": "BUILD_AG", "com": "BUILD_COM", "def": "BUILD_DEF"}
		cmdType := typeMap[strings.ToLower(parts[2])]
		if cmdType == "" {
			m.feedback = "unknown type — use ag, com, or def"
			return m
		}
		ok := m.engine.QueueCommand(models.Command{
			Type:   cmdType,
			Params: map[string]string{"city": parts[1]},
			Cost:   2,
		})
		if ok {
			m.feedback = "queued " + cmdType + " on " + parts[1]
		} else {
			m.feedback = "not enough CP"
		}
	default:
		m.feedback = "unknown command: " + parts[0]
	}
	return m
}

func tailLogs(logs []models.LogEntry, n int) []models.LogEntry {
	if len(logs) <= n {
		cp := make([]models.LogEntry, len(logs))
		copy(cp, logs)
		return cp
	}
	cp := make([]models.LogEntry, n)
	copy(cp, logs[len(logs)-n:])
	return cp
}
