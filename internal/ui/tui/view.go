package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")) // gold
	styleSeason   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114")) // green
	styleElement  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))  // cyan
	styleCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")) // orange
	styleFeedback = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))            // red
	styleGood     = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))            // green
	styleDim         = lipgloss.NewStyle().Faint(true)
	styleFooter      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))            // mid-grey
	styleSplashMtn   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("24")) // dark slate-blue (distant mountains)
	styleSplashTitle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("136")) // dark amber (dim gold)
)

var seasonDisplay = map[string][2]string{
	"Spring":     {"~~~ spring ~~~", "木"},
	"Summer":     {"=== summer ===", "火"},
	"Autumn":     {"::: autumn :::", "金"},
	"Winter":     {"... winter ...", "水"},
	"Transition": {"--- transition -", "土"},
}

const banner = "帝王战纪：三国录  —  Sovereign Record"
const bannerChinese = "帝王战纪：三国录"

const (
	padTop    = 3
	padBottom = 3
	padLeft   = 3
	padRight  = 3
)

func (m model) dividerLine() string {
	w := m.width - padLeft - padRight
	if w < 20 {
		w = 40
	}
	return strings.Repeat("─", w)
}

func indentLines(s string) string {
	prefix := strings.Repeat(" ", padLeft)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func (m model) mapBody(body string) string {
	season := "Spring"
	if m.engine != nil {
		season = m.engine.GetState().Season
	}
	cityPhase := m.tickCount / 20
	wavePhase := m.tickCount / 25
	framedW := mapW + 2 // +2 for left/right border chars
	rightW := m.width - framedW - 4 - padLeft - padRight
	if rightW < 20 {
		rightW = 40
	}
	cities := m.ledger.SortedCities()
	selectedCity := ""
	if m.screen == screenGameB && len(cities) > 0 {
		selectedCity = cities[m.cityCursor].Name
	}
	return joinColumns(CachedRenderMap(cities, cityPhase, season, wavePhase, selectedCity, m.queuedCities), body, framedW, 4, rightW)
}

// joinColumns places left and right strings side by side, padding left to fixedW runes.
// rightW > 0 wraps the right column at that width.
func joinColumns(left, right string, fixedW, gap, rightW int) string {
	if rightW > 0 {
		right = wrapText(right, rightW)
	}
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

// wrapText word-wraps plain text to width. Styled (ANSI) lines are passed through unchanged.
func wrapText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var out strings.Builder
	for i, line := range lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		if lipgloss.Width(line) <= width {
			out.WriteString(line)
			continue
		}
		words := strings.Fields(line)
		col := 0
		for j, word := range words {
			wl := lipgloss.Width(word)
			if j == 0 {
				out.WriteString(word)
				col = wl
			} else if col+1+wl > width {
				out.WriteByte('\n')
				out.WriteString(word)
				col = wl
			} else {
				out.WriteByte(' ')
				out.WriteString(word)
				col += 1 + wl
			}
		}
	}
	return out.String()
}

func (m model) headerSimple() string {
	return styleTitle.Render(banner) + "\n" + m.dividerLine() + "\n\n"
}

// gameStatus returns the date/season/resource line rendered into the right pane, above the cycle label.
func (m model) gameStatus() string {
	state := m.engine.GetState()
	var season string
	if sd, ok := seasonDisplay[state.Season]; ok {
		season = styleSeason.Render(sd[0]) + "  " + styleElement.Render(sd[1]) + "  " + styleDim.Render(strings.ToLower(state.Element))
	}
	dateLine := fmt.Sprintf("[ %d.%02d ]  %s", state.Year, state.Month, season)
	resLine := fmt.Sprintf("grain: %d   gold: %d   CP: %d",
		state.Resources.Grain, state.Resources.Gold, state.AvailableCP)
	return dateLine + "\n" + styleDim.Render(resLine) + "\n\n"
}

func (m model) footer() string {
	if m.showHelp {
		return m.dividerLine() + "\n" + styleFooter.Render("  [ ? ] close help   [ q ] quit")
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
		hints = "[ enter ] continue   [ q ] quit   [ ? ] help"
	case screenGameB:
		if m.showLog {
			hints = "[ l ] close log   [ q ] quit   [ ? ] help"
		} else {
			hints = "[ ↑↓ ] city   [ a ] ag   [ c ] com   [ d ] def   [ l ] log   [ x ] end turn   [ q ] quit   [ ? ] help"
		}
	case screenGameC:
		hints = "[ enter ] next month   [ q ] quit   [ ? ] help"
	}
	return m.dividerLine() + "\n" + styleFooter.Render("  "+hints)
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

// splashCanvas is a small parameterised braille pixel canvas for splash art.
type splashCanvas struct {
	pw, ph int
	dots   []bool
}

func newSplashCanvas(pw, ph int) *splashCanvas {
	return &splashCanvas{pw: pw, ph: ph, dots: make([]bool, pw*ph)}
}

func (c *splashCanvas) set(x, y int) {
	if x >= 0 && x < c.pw && y >= 0 && y < c.ph {
		c.dots[y*c.pw+x] = true
	}
}

func (c *splashCanvas) render() string {
	bw, bh := c.pw/2, c.ph/4
	var sb strings.Builder
	for row := 0; row < bh; row++ {
		for col := 0; col < bw; col++ {
			px, py := col*2, row*4
			var b rune
			if c.dots[(py+0)*c.pw+(px+0)] { b |= 0x01 }
			if c.dots[(py+1)*c.pw+(px+0)] { b |= 0x02 }
			if c.dots[(py+2)*c.pw+(px+0)] { b |= 0x04 }
			if c.dots[(py+3)*c.pw+(px+0)] { b |= 0x40 }
			if c.dots[(py+0)*c.pw+(px+1)] { b |= 0x08 }
			if c.dots[(py+1)*c.pw+(px+1)] { b |= 0x10 }
			if c.dots[(py+2)*c.pw+(px+1)] { b |= 0x20 }
			if c.dots[(py+3)*c.pw+(px+1)] { b |= 0x80 }
			sb.WriteRune(0x2800 | b)
		}
		if row < bh-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// pixelGlyph is a 5×7 (or 3×7 for space) bitmap glyph; bit (w-1) = leftmost col.
type pixelGlyph struct {
	w    int
	rows [7]uint8
}

var pixelFont = map[rune]pixelGlyph{
	'S': {5, [7]uint8{0b01110, 0b10000, 0b10000, 0b01110, 0b00001, 0b00001, 0b01110}},
	'O': {5, [7]uint8{0b01110, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b01110}},
	'V': {5, [7]uint8{0b10001, 0b10001, 0b10001, 0b01010, 0b01010, 0b00100, 0b00100}},
	'E': {5, [7]uint8{0b11111, 0b10000, 0b10000, 0b11110, 0b10000, 0b10000, 0b11111}},
	'R': {5, [7]uint8{0b11110, 0b10001, 0b10001, 0b11110, 0b10100, 0b10010, 0b10001}},
	'I': {5, [7]uint8{0b01110, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100, 0b01110}},
	'G': {5, [7]uint8{0b01110, 0b10000, 0b10000, 0b10011, 0b10001, 0b10001, 0b01110}},
	'N': {5, [7]uint8{0b10001, 0b11001, 0b11001, 0b10101, 0b10011, 0b10011, 0b10001}},
	'C': {5, [7]uint8{0b01110, 0b10000, 0b10000, 0b10000, 0b10000, 0b10000, 0b01110}},
	'D': {5, [7]uint8{0b11110, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b11110}},
	'T': {5, [7]uint8{0b11111, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100}},
	'H': {5, [7]uint8{0b10001, 0b10001, 0b10001, 0b11111, 0b10001, 0b10001, 0b10001}},
	'K': {5, [7]uint8{0b10001, 0b10010, 0b10100, 0b11000, 0b10100, 0b10010, 0b10001}},
	'M': {5, [7]uint8{0b10001, 0b11011, 0b10101, 0b10001, 0b10001, 0b10001, 0b10001}},
	' ': {3, [7]uint8{0, 0, 0, 0, 0, 0, 0}},
}

var (
	splashArtOnce    sync.Once
	splashArtStr     string
	titleArtOnce     sync.Once
	titleArtStr      string
	subtitleArtOnce  sync.Once
	subtitleArtStr   string
)

func getSplashArt() string {
	splashArtOnce.Do(func() { splashArtStr = makeSplashArt() })
	return splashArtStr
}

func getTitleArt() string {
	titleArtOnce.Do(func() { titleArtStr = renderBrailleText("SOVEREIGN RECORD") })
	return titleArtStr
}

func getSubtitleArt() string {
	subtitleArtOnce.Do(func() { subtitleArtStr = renderBrailleText("THREE KINGDOMS") })
	return subtitleArtStr
}

// renderBrailleText renders text using the 5×7 pixel font into a 160px-wide braille canvas, centred.
// Each glyph row is drawn twice (2× height scale) to produce 4 braille rows of visible text.
func renderBrailleText(text string) string {
	runes := []rune(text)
	totalPx := 0
	for i, r := range runes {
		g := pixelFont[r]
		totalPx += g.w
		if i < len(runes)-1 {
			totalPx++
		}
	}
	const spw, sph = 160, 16
	c := newSplashCanvas(spw, sph)
	xOff := (spw - totalPx) / 2
	if xOff < 0 {
		xOff = 0
	}
	const yOff = 2 // shift off braille-cell boundary to avoid crossbar artefact
	x := xOff
	for i, r := range runes {
		g := pixelFont[r]
		for row := 0; row < 7; row++ {
			py := row*2 + yOff
			for col := 0; col < g.w; col++ {
				if g.rows[row]&(1<<uint(g.w-1-col)) != 0 {
					c.set(x+col, py)
					c.set(x+col, py+1)
				}
			}
		}
		x += g.w
		if i < len(runes)-1 {
			x++
		}
	}
	return c.render()
}

func makeSplashArt() string {
	const spw, sph = 160, 52
	c := newSplashCanvas(spw, sph)
	base := sph - 1

	type peakDef struct{ cx, cy, hw int }
	peaks := []peakDef{{16, 6, 20}, {80, 0, 29}, {141, 7, 17}}

	topAt := func(x int) int {
		t := base + 1
		for _, p := range peaks {
			dx := x - p.cx
			if dx < 0 {
				dx = -dx
			}
			if dx > p.hw {
				continue
			}
			y := p.cy + dx*(base-p.cy)/p.hw
			if y < t {
				t = y
			}
		}
		return t
	}

	// filled mountain silhouettes
	for x := 0; x < spw; x++ {
		for y := topAt(x); y <= base; y++ {
			c.set(x, y)
		}
	}

	// ground line in valleys between peaks
	for x := 0; x < spw; x++ {
		if topAt(x) > base {
			c.set(x, base)
		}
	}

	// trees: triangular crown above surface
	addTree := func(tx int) {
		s := topAt(tx)
		if s > base {
			s = base
		}
		c.set(tx, s-10)
		for dx := -1; dx <= 1; dx++ { c.set(tx+dx, s-9) }
		for dx := -2; dx <= 2; dx++ { c.set(tx+dx, s-8) }
		for dx := -3; dx <= 3; dx++ { c.set(tx+dx, s-7) }
		for dx := -4; dx <= 4; dx++ { c.set(tx+dx, s-6) }
		c.set(tx, s-5)
		c.set(tx, s-4)
		c.set(tx, s-3)
	}
	// valley 1: x≈37–50 (between peak 1 and peak 2)
	for _, tx := range []int{38, 43, 48} {
		addTree(tx)
	}
	// valley 2: x≈110–123 (between peak 2 and peak 3)
	for _, tx := range []int{111, 116, 121} {
		addTree(tx)
	}

	// river through valley 1: thin meandering line
	for i := 0; i < 14; i++ {
		rx := 37 + i
		s := topAt(rx)
		if s > base {
			s = base
		}
		c.set(rx, s-2)
		if i%3 != 1 {
			c.set(rx, s-3)
		}
	}

	return c.render()
}

var shineStyles = [4]lipgloss.Style{
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("222")),
}

// bigTitleShine renders the Chinese banner with wide character spacing and a left-to-right shine sweep.
func (m model) bigTitleShine() string {
	runes := []rune(bannerChinese)
	pos := (m.tickCount / 3) % (len(runes) + 20)
	var sb strings.Builder
	for i, r := range runes {
		d := pos - i
		if d >= 0 && d < len(shineStyles) {
			sb.WriteString(shineStyles[d].Render(string(r)))
		} else {
			sb.WriteString(styleTitle.Render(string(r)))
		}
		sb.WriteRune(' ')
	}
	return strings.TrimRight(sb.String(), " ") + styleDim.Render("  |  teio senki")
}

func (m model) viewSplash() string {
	art := styleSplashMtn.Render(getSplashArt())
	titleArt := styleTitle.Render(getTitleArt())
	subtitleArt := styleSplashTitle.Render(getSubtitleArt())
	animated := styleTitle.Render(string(splashFull[:m.charIdx]))
	return m.headerSimple() +
		"\n" +
		m.bigTitleShine() + "\n" +
		titleArt + "\n" +
		subtitleArt + "\n\n" +
		art + "\n\n" +
		animated
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
	b.WriteString(m.gameStatus())
	b.WriteString(styleSeason.Render("world update") + "\n\n")
	for _, e := range m.cycleALogs {
		fmt.Fprintf(&b, "  %s %s\n", styleDim.Render("["+e.Type+"]"), e.Description)
	}
	return m.headerSimple() + m.mapBody(b.String())
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
	fmt.Fprintf(&b, "%s\n\n", styleSeason.Render("orders"))
	b.WriteString("  ↑ / k        previous city\n")
	b.WriteString("  ↓ / j        next city\n")
	b.WriteString("  a            queue build agriculture\n")
	b.WriteString("  c            queue build commerce\n")
	b.WriteString("  d            queue build defense\n")
	b.WriteString("  x            end turn (settle)\n\n")
	return b.String()
}

func (m model) viewCycleB() string {
	var b strings.Builder
	b.WriteString(m.gameStatus())
	if m.showLog {
		b.WriteString(styleSeason.Render("log") + "\n\n")
		logs := tailLogs(m.ledger.Logs, 50)
		for _, e := range logs {
			fmt.Fprintf(&b, "  %s %s\n", styleDim.Render("["+e.Type+"]"), e.Description)
		}
		return m.headerSimple() + m.mapBody(b.String())
	}
	b.WriteString(styleSeason.Render("orders") + "\n\n")
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
		queued := ""
		if m.queuedCities[cities[i].Name] {
			queued = "  " + styleCursor.Render("*")
		}
		if m.cityCursor == i {
			fmt.Fprintf(&b, "%s%-20s  %4d  %4d  %4d%s\n",
				styleCursor.Render("> "), cities[i].Name, cities[i].Agriculture, cities[i].Commerce, cities[i].Defense, queued)
		} else {
			fmt.Fprintf(&b, "  %-20s  %4d  %4d  %4d%s\n",
				cities[i].Name, cities[i].Agriculture, cities[i].Commerce, cities[i].Defense, queued)
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
	return m.headerSimple() + m.mapBody(b.String())
}

func (m model) viewCycleC() string {
	var b strings.Builder
	b.WriteString(m.gameStatus())
	b.WriteString(styleSeason.Render("settlement") + "\n\n")
	for _, e := range m.cycleCLogs {
		fmt.Fprintf(&b, "  %s %s\n", styleDim.Render("["+e.Type+"]"), e.Description)
	}
	return m.headerSimple() + m.mapBody(b.String())
}
