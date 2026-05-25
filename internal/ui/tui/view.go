package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")) // gold
	styleSeason   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114")) // green
	styleElement  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))  // cyan
	styleCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")) // orange
	styleFeedback = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))            // red
	styleGood     = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))            // green
	styleDim      = lipgloss.NewStyle().Faint(true)
	styleFooter   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))            // mid-grey
)

var seasonDisplay = map[string][2]string{
	"Spring":     {"~~~ spring ~~~", "木"},
	"Summer":     {"=== summer ===", "火"},
	"Autumn":     {"::: autumn :::", "金"},
	"Winter":     {"... winter ...", "水"},
	"Transition": {"--- transition -", "土"},
}

const banner = "帝王战纪：三国录  —  Sovereign Record"
const divider = "────────────────────────────────────────"

const (
	padTop    = 5
	padBottom = 5
	padLeft   = 3
)

func indentLines(s string) string {
	prefix := strings.Repeat(" ", padLeft)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func (m model) mapBody(body string) string {
	return joinColumns(RenderMap(m.ledger.SortedCities()), body, mapW, 4)
}

// joinColumns places left and right strings side by side, padding left to fixedW runes.
func joinColumns(left, right string, fixedW, gap int) string {
	ll := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rl := strings.Split(strings.TrimRight(right, "\n"), "\n")
	h := len(ll)
	if len(rl) > h {
		h = len(rl)
	}
	sep := strings.Repeat(" ", gap)
	var sb strings.Builder
	for i := range h {
		var l, r string
		if i < len(ll) {
			l = ll[i]
		}
		if i < len(rl) {
			r = rl[i]
		}
		pad := fixedW - lipgloss.Width(l)
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(styleDim.Render(l))
		sb.WriteString(strings.Repeat(" ", pad))
		sb.WriteString(sep)
		sb.WriteString(r)
		if i < h-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func (m model) headerSimple() string {
	return styleTitle.Render(banner) + "\n" + divider + "\n\n"
}

func (m model) headerGame() string {
	state := m.engine.GetState()
	var season string
	if sd, ok := seasonDisplay[state.Season]; ok {
		season = styleSeason.Render(sd[0]) + "  " + styleElement.Render(sd[1]) + "  " + styleDim.Render(strings.ToLower(state.Element))
	}
	statusLine := fmt.Sprintf("[ %d.%02d ]  %s     grain: %d   gold: %d   CP: %d",
		state.Year, state.Month, season,
		state.Resources.Grain, state.Resources.Gold, state.AvailableCP)
	return styleTitle.Render(banner) + "\n" + statusLine + "\n" + divider + "\n\n"
}

func (m model) footer() string {
	if m.showHelp {
		return divider + "\n" + styleFooter.Render("  [ ? ] close help   [ q ] quit")
	}
	var hints string
	switch m.screen {
	case screenSplash:
		hints = "[ enter ] continue"
	case screenMenu:
		hints = "[ ↑↓ ] navigate   [ enter ] select   [ q ] quit"
	case screenScenario, screenSovereign:
		hints = "[ ↑↓ ] navigate   [ enter ] select   [ q ] quit   [ ? ] help"
	case screenBriefing:
		hints = "[ enter ] begin   [ q ] quit   [ ? ] help"
	case screenGameA:
		hints = "[ enter ] issue commands   [ q ] quit   [ ? ] help"
	case screenGameB:
		hints = "[ ↑↓ ] city   [ a ] ag   [ c ] com   [ d ] def   [ x ] end turn   [ q ] quit   [ ? ] help"
	case screenGameC:
		hints = "[ enter ] next month   [ q ] quit   [ ? ] help"
	}
	return divider + "\n" + styleFooter.Render("  "+hints)
}

func (m model) withFooter(content string) string {
	top := strings.Repeat("\n", padTop)
	indented := indentLines(content)
	footer := indentLines(m.footer())
	if m.height == 0 {
		return top + indented + "\n" + footer
	}
	// rows: padTop + contentRows + 1(sep) + 2(footer) + padBottom = m.height
	contentRows := strings.Count(indented, "\n") + 1
	pad := m.height - padTop - padBottom - 3 - contentRows
	if pad > 0 {
		indented += strings.Repeat("\n", pad)
	}
	return top + indented + "\n" + footer + strings.Repeat("\n", padBottom)
}

func (m model) View() string {
	if m.showHelp {
		return m.withFooter(m.viewHelp())
	}
	switch m.screen {
	case screenSplash:
		return m.withFooter(m.viewSplash())
	case screenMenu:
		return m.withFooter(m.viewMenu())
	case screenScenario:
		return m.withFooter(m.viewScenario())
	case screenSovereign:
		return m.withFooter(m.viewSovereign())
	case screenBriefing:
		return m.withFooter(m.viewBriefing())
	case screenGameA:
		return m.withFooter(m.viewCycleA())
	case screenGameB:
		return m.withFooter(m.viewCycleB())
	case screenGameC:
		return m.withFooter(m.viewCycleC())
	}
	return ""
}

func (m model) viewSplash() string {
	// splashFull contains plain text; colour only the rendered banner portion
	text := string(splashFull[:m.charIdx])
	return strings.Replace(text, banner, styleTitle.Render(banner), 1)
}

func (m model) viewMenu() string {
	sel := func(i int) string {
		if m.cursor == i {
			return styleCursor.Render(">") + " "
		}
		return "  "
	}
	return m.headerSimple() +
		sel(0) + "new game\n" +
		"    load game  " + styleDim.Render("[coming soon]") + "\n" +
		sel(1) + "quit\n"
}

func (m model) viewScenario() string {
	var b strings.Builder
	b.WriteString(styleSeason.Render("select scenario") + "\n\n")
	for i, s := range allScenarios {
		cur := "  "
		if m.cursor == i {
			cur = styleCursor.Render(">") + " "
		}
		lock := ""
		if !s.unlocked {
			lock = "  " + styleDim.Render("[locked]")
		}
		fmt.Fprintf(&b, "%s%s  %s%s\n", cur, s.epoch, s.name, lock)
	}
	return m.headerSimple() + m.mapBody(b.String())
}

func (m model) viewSovereign() string {
	var b strings.Builder
	sc := allScenarios[m.scenarioIdx]
	fmt.Fprintf(&b, "%s  —  %s\n\n", styleSeason.Render("select sovereign"), sc.name)
	fmt.Fprintf(&b, "  %-22s %-8s  STR  VAL  GOV\n", "name", "essence")
	fmt.Fprintf(&b, "  %s\n", strings.Repeat("-", 50))

	const pageSize = 10
	pageStart := (m.cursor / pageSize) * pageSize
	pageEnd := pageStart + pageSize
	if pageEnd > len(m.lords) {
		pageEnd = len(m.lords)
	}
	for i := pageStart; i < pageEnd; i++ {
		lord := m.lords[i]
		if m.cursor == i {
			fmt.Fprintf(&b, "%s%-22s %-8s  %3d  %3d  %3d\n",
				styleCursor.Render(">")+" ", lord.Name, lord.Essence, lord.Strategy, lord.Valour, lord.Governance)
		} else {
			fmt.Fprintf(&b, "  %-22s %-8s  %3d  %3d  %3d\n",
				lord.Name, lord.Essence, lord.Strategy, lord.Valour, lord.Governance)
		}
	}
	if len(m.lords) > pageSize {
		fmt.Fprintf(&b, "  page %d/%d\n", m.cursor/pageSize+1, (len(m.lords)+pageSize-1)/pageSize)
	}
	return m.headerSimple() + m.mapBody(b.String())
}

func (m model) viewBriefing() string {
	sc := allScenarios[m.scenarioIdx]
	var body strings.Builder
	fmt.Fprintf(&body, "%s  —  %s\n\n", styleSeason.Render(sc.name), sc.epoch)
	fmt.Fprintf(&body, "%s\n\n", sc.desc)
	if lord, ok := m.ledger.GetOfficer(m.chosenLord); ok {
		fmt.Fprintf(&body, "sovereign  : %s", styleTitle.Render(lord.Name))
		if lord.Title != "" {
			fmt.Fprintf(&body, " %s", styleDim.Render("("+lord.Title+")"))
		}
		fmt.Fprintf(&body, "\nessence    : %s\n", styleElement.Render(lord.Essence))
		fmt.Fprintf(&body, "strategy   : %d   valour: %d   governance: %d\n",
			lord.Strategy, lord.Valour, lord.Governance)
	}
	return m.headerSimple() + m.mapBody(body.String())
}

func (m model) viewCycleA() string {
	var b strings.Builder
	b.WriteString(styleSeason.Render("cycle A — world update") + "\n\n")
	for _, e := range m.cycleALogs {
		fmt.Fprintf(&b, "  %s %s\n", styleDim.Render("["+e.Type+"]"), e.Description)
	}
	return m.headerGame() + m.mapBody(b.String())
}

func (m model) viewHelp() string {
	var b strings.Builder
	b.WriteString(m.headerSimple())
	fmt.Fprintf(&b, "%s\n\n", styleTitle.Render("help — key bindings"))
	fmt.Fprintf(&b, "%s\n\n", styleSeason.Render("global"))
	b.WriteString("  q / ctrl+c   quit\n")
	b.WriteString("  ?            toggle this help\n\n")
	fmt.Fprintf(&b, "%s\n\n", styleSeason.Render("navigation (menus)"))
	b.WriteString("  ↑ / k        move up\n")
	b.WriteString("  ↓ / j        move down\n")
	b.WriteString("  enter        confirm\n\n")
	fmt.Fprintf(&b, "%s\n\n", styleSeason.Render("cycle B — commands"))
	b.WriteString("  ↑ / k        previous city\n")
	b.WriteString("  ↓ / j        next city\n")
	b.WriteString("  a            queue build agriculture\n")
	b.WriteString("  c            queue build commerce\n")
	b.WriteString("  d            queue build defense\n")
	b.WriteString("  x / enter    end turn (settle)\n\n")
	return b.String()
}

func (m model) viewCycleB() string {
	var b strings.Builder
	b.WriteString(styleSeason.Render("cycle B — commands") + "\n\n")
	const pageSize = 10
	fmt.Fprintf(&b, "  %-20s  %4s  %4s  %4s\n", "city", "ag", "com", "def")
	fmt.Fprintf(&b, "  %s\n", strings.Repeat("-", 38))
	cities := m.ledger.SortedCities()
	pageStart := (m.cityCursor / pageSize) * pageSize
	pageEnd := pageStart + pageSize
	if pageEnd > len(cities) {
		pageEnd = len(cities)
	}
	for i := pageStart; i < pageEnd; i++ {
		if m.cityCursor == i {
			fmt.Fprintf(&b, "%s%-20s  %4d  %4d  %4d\n",
				styleCursor.Render("> "), cities[i].Name, cities[i].Agriculture, cities[i].Commerce, cities[i].Defense)
		} else {
			fmt.Fprintf(&b, "  %-20s  %4d  %4d  %4d\n",
				cities[i].Name, cities[i].Agriculture, cities[i].Commerce, cities[i].Defense)
		}
	}
	if len(cities) > pageSize {
		fmt.Fprintf(&b, "  page %d/%d\n", m.cityCursor/pageSize+1, (len(cities)+pageSize-1)/pageSize)
	}
	if m.feedback != "" {
		if strings.HasPrefix(m.feedback, "not enough") {
			fmt.Fprintf(&b, "\n  %s\n", styleFeedback.Render(m.feedback))
		} else {
			fmt.Fprintf(&b, "\n  %s\n", styleGood.Render(m.feedback))
		}
	}
	return m.headerGame() + m.mapBody(b.String())
}

func (m model) viewCycleC() string {
	var b strings.Builder
	b.WriteString(styleSeason.Render("cycle C — settlement") + "\n\n")
	for _, e := range m.cycleCLogs {
		fmt.Fprintf(&b, "  %s %s\n", styleDim.Render("["+e.Type+"]"), e.Description)
	}
	return m.headerGame() + m.mapBody(b.String())
}
