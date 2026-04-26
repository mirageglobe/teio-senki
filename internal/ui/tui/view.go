package tui

import (
	"fmt"
	"strings"
)

const banner = "帝王战纪：三国录  —  Sovereign Record"
const divider = "────────────────────────────────────────"

func (m model) View() string {
	switch m.screen {
	case screenSplash:
		return m.viewSplash()
	case screenMenu:
		return m.viewMenu()
	case screenScenario:
		return m.viewScenario()
	case screenSovereign:
		return m.viewSovereign()
	case screenBriefing:
		return m.viewBriefing()
	case screenGameA:
		return m.viewCycleA()
	case screenGameB:
		return m.viewCycleB()
	case screenGameC:
		return m.viewCycleC()
	}
	return ""
}

func (m model) viewSplash() string {
	return fmt.Sprintf("%s\n%s\n\n%s\n\n[ press enter ]\n",
		banner, divider,
		`"Sovereignty through the Ledger, Strategy through the Elements."`)
}

func (m model) viewMenu() string {
	sel := func(i int) string {
		if m.cursor == i {
			return "> "
		}
		return "  "
	}
	return banner + "\n\n" +
		sel(0) + "new game\n" +
		"    load game  [coming soon]\n" +
		sel(1) + "quit\n"
}

func (m model) viewScenario() string {
	var b strings.Builder
	fmt.Fprintf(&b, "select scenario\n%s\n\n", divider)
	for i, s := range allScenarios {
		cur := "  "
		if m.cursor == i {
			cur = "> "
		}
		lock := ""
		if !s.unlocked {
			lock = "  [locked]"
		}
		fmt.Fprintf(&b, "%s%s  %s%s\n", cur, s.epoch, s.name, lock)
	}
	b.WriteString("\n[↑↓] navigate  [enter] select\n")
	return b.String()
}

func (m model) viewSovereign() string {
	var b strings.Builder
	sc := allScenarios[m.scenarioIdx]
	fmt.Fprintf(&b, "select sovereign  —  %s\n%s\n\n", sc.name, divider)
	fmt.Fprintf(&b, "  %-22s %-8s  STR  VAL  GOV\n", "name", "essence")
	fmt.Fprintf(&b, "  %s\n", strings.Repeat("-", 50))

	visibleHeight := m.height - 7
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	end := m.scrollOffset + visibleHeight
	if end > len(m.lords) {
		end = len(m.lords)
	}
	for i := m.scrollOffset; i < end; i++ {
		lord := m.lords[i]
		cur := "  "
		if m.cursor == i {
			cur = "> "
		}
		fmt.Fprintf(&b, "%s%-22s %-8s  %3d  %3d  %3d\n",
			cur, lord.Name, lord.Essence, lord.Strategy, lord.Valour, lord.Governance)
	}
	b.WriteString("\n[↑↓] navigate  [enter] confirm\n")
	return b.String()
}

func (m model) viewBriefing() string {
	var b strings.Builder
	sc := allScenarios[m.scenarioIdx]
	fmt.Fprintf(&b, "%s  —  %s\n%s\n\n", sc.name, sc.epoch, divider)
	fmt.Fprintf(&b, "%s\n\n", sc.desc)
	if lord, ok := m.ledger.GetOfficer(m.chosenLord); ok {
		fmt.Fprintf(&b, "sovereign  : %s", lord.Name)
		if lord.Title != "" {
			fmt.Fprintf(&b, " (%s)", lord.Title)
		}
		fmt.Fprintf(&b, "\nessence    : %s\n", lord.Essence)
		fmt.Fprintf(&b, "strategy   : %d   valour: %d   governance: %d\n",
			lord.Strategy, lord.Valour, lord.Governance)
	}
	b.WriteString("\n[ press enter to begin ]\n")
	return b.String()
}

func (m model) viewCycleA() string {
	state := m.engine.GetState()
	var b strings.Builder
	fmt.Fprintf(&b, "[ %d.%02d ]  cycle A — world update\n%s\n\n", state.Year, state.Month, divider)
	for _, e := range m.cycleALogs {
		fmt.Fprintf(&b, "  [%s] %s\n", e.Type, e.Description)
	}
	b.WriteString("\n[ enter ] issue commands\n")
	return b.String()
}

func (m model) viewCycleB() string {
	state := m.engine.GetState()
	var b strings.Builder
	fmt.Fprintf(&b, "[ %d.%02d ]  cycle B — commands  (CP: %d)\n%s\n\n",
		state.Year, state.Month, state.AvailableCP, divider)
	b.WriteString("  build <city> ag|com|def   end\n\n")
	if m.feedback != "" {
		fmt.Fprintf(&b, "  %s\n\n", m.feedback)
	}
	fmt.Fprintf(&b, "> %s_\n", m.input)
	return b.String()
}

func (m model) viewCycleC() string {
	state := m.engine.GetState()
	var b strings.Builder
	fmt.Fprintf(&b, "[ %d.%02d ]  cycle C — settlement\n%s\n\n", state.Year, state.Month, divider)
	for _, e := range m.cycleCLogs {
		fmt.Fprintf(&b, "  [%s] %s\n", e.Type, e.Description)
	}
	b.WriteString("\n[ enter ] next month\n")
	return b.String()
}
