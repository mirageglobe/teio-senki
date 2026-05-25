// Package tui owns the terminal UI frontend. It is a dumb view: it calls engine
// methods and renders text. It does NOT contain game logic.
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/engine/sovereign"
	"github.com/mirageglobe/teio-senki/internal/models"
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(35*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

var splashFull = []rune(banner + "\n" + divider + "\n\n" +
	`"Sovereignty through the Ledger, Strategy through the Elements."` +
	"\n\n[ press enter ]\n")

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
	screen       screen
	ledger       *ledger.Ledger
	engine       *sovereign.Engine
	cursor       int
	scrollOffset int
	scenarioIdx  int
	lords        []models.Officer
	chosenLord   string
	cycleALogs   []models.LogEntry
	cycleCLogs   []models.LogEntry
	cityCursor   int
	feedback     string
	width        int
	height       int
	charIdx      int
	showHelp     bool
	tickCount    int
	mapPulse     bool
}

// Run initialises the Bubble Tea program and blocks until the player exits.
func Run(l *ledger.Ledger) error {
	m := model{ledger: l, lords: l.Lords()}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		m.tickCount++
		m.mapPulse = (m.tickCount/15)%2 == 0 // ~500 ms half-cycle at 35 ms/tick
		if m.screen == screenSplash && m.charIdx < len(splashFull) {
			m.charIdx++
		}
		return m, tick()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenSplash:
		if key == "enter" || key == " " {
			if m.charIdx < len(splashFull) {
				m.charIdx = len(splashFull)
			} else {
				m.screen, m.cursor = screenMenu, 0
			}
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
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lords)-1 {
				m.cursor++
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
			m.cityCursor = 0
			m.screen = screenGameB
		}
	case screenGameB:
		cities := m.ledger.SortedCities()
		switch key {
		case "up", "k":
			if m.cityCursor > 0 {
				m.cityCursor--
			}
		case "down", "j":
			if m.cityCursor < len(cities)-1 {
				m.cityCursor++
			}
		case "a", "c", "d":
			if len(cities) == 0 {
				return m, nil
			}
			city := cities[m.cityCursor]
			typeMap := map[string]string{"a": "BUILD_AG", "c": "BUILD_COM", "d": "BUILD_DEF"}
			ok := m.engine.QueueCommand(models.Command{
				Type:   typeMap[key],
				Params: map[string]string{"city": city.Name},
				Cost:   2,
			})
			if ok {
				m.feedback = "queued " + typeMap[key] + " → " + city.Name
			} else {
				m.feedback = "not enough CP"
			}
		case "x", "enter":
			m.cycleCLogs = m.engine.SettleTurn()
			m.feedback = ""
			m.screen = screenGameC
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
