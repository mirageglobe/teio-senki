package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mirageglobe/teio-senki/internal/models"
)

// City pulse gradient: bold; index 0 = brightest, 5 = dimmest; ping-pong for fluid glow.
// Range stays bright yellow→white — never dips into terrain brown/tan territory.
var mapStyleCity = [6]lipgloss.Style{
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("253")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")),
	lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228")),
}

var (
	mapStyleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201")) // bright magenta — selected city
	mapStyleWave     = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))              // ocean blue
	mapStyleMountain = lipgloss.NewStyle().Foreground(lipgloss.Color("130"))             // brown
	mapStyleHills    = lipgloss.NewStyle().Foreground(lipgloss.Color("179")) // tan
	mapStyleForest   = lipgloss.NewStyle().Foreground(lipgloss.Color("107")) // light sage green
	mapStyleRiver    = lipgloss.NewStyle().Foreground(lipgloss.Color("38"))  // cyan-blue
	mapStyleMarsh    = lipgloss.NewStyle().Foreground(lipgloss.Color("66"))  // muted teal
	mapStyleSteppe   = lipgloss.NewStyle().Foreground(lipgloss.Color("143")) // olive
)

var seasonBorderColour = map[string]lipgloss.Color{
	"Spring":     "46",  // bright green  (Wood)
	"Summer":     "196", // bright red    (Fire)
	"Autumn":     "220", // bright gold   (Metal)
	"Winter":     "51",  // bright cyan   (Water)
	"Transition": "172", // bright amber  (Earth)
}

const mapW, mapH = 60, 23 // braille chars; pixel grid = 120×92
const pw, ph = mapW * 2, mapH * 4

const lonMin, lonMax = 95.0, 136.0
const latMin, latMax = 18.0, 50.0

// xStretch widens the projection to compensate for terminal font aspect ratio (chars ~2× taller than wide).
// 1.0 = full 73–136° visible (shows Japan/Taiwan); higher values narrow the visible range.
const xStretch = 1.0

// Simplified China border polygon (lon, lat), clockwise.
var chinaBorder = [][2]float64{
	{135.0, 48.5}, {132.0, 47.5}, {130.5, 42.0}, {129.0, 42.5},
	{128.5, 44.0}, {124.5, 40.0}, {121.5, 38.5}, {122.0, 37.0},
	{120.5, 36.0}, {119.5, 32.0}, {121.5, 29.0}, {121.0, 27.0},
	{118.0, 24.5}, {114.5, 22.5}, {113.0, 22.0}, {110.5, 20.5},
	{109.5, 21.5}, {108.5, 22.0}, {106.5, 22.5}, {103.5, 22.0},
	{101.5, 22.5}, {100.0, 22.0}, {98.5, 23.5}, {97.5, 25.0},
	{97.0, 28.0}, {96.5, 29.0}, {92.0, 28.0}, {88.5, 27.5},
	{84.0, 28.5}, {80.5, 30.0}, {78.5, 32.0}, {76.5, 35.0},
	{74.0, 37.0}, {73.5, 39.5}, {74.5, 40.5}, {76.0, 41.5},
	{80.0, 42.0}, {82.5, 45.0}, {87.0, 48.0}, {89.0, 48.5},
	{97.0, 49.5}, {106.5, 50.0}, {110.0, 53.0}, {119.5, 53.5},
	{122.5, 52.5}, {126.0, 52.0}, {128.0, 50.0}, {130.5, 48.0},
	{135.0, 48.5},
}

var taiwanBorder = [][2]float64{
	{121.5, 25.3}, {122.0, 24.5}, {121.6, 23.0}, {120.8, 22.0},
	{120.2, 22.8}, {120.4, 23.7}, {121.0, 24.5}, {121.5, 25.3},
}

var kyushuBorder = [][2]float64{
	{130.2, 33.6}, {130.8, 33.9}, {131.2, 33.5}, {131.7, 32.5},
	{131.4, 31.6}, {130.6, 31.5}, {130.0, 32.2}, {129.8, 32.9},
	{130.2, 33.6},
}

var shikokuBorder = [][2]float64{
	{132.6, 34.2}, {133.6, 34.3}, {134.5, 33.8}, {133.8, 33.5},
	{132.8, 33.5}, {132.6, 34.2},
}

var honshuBorder = [][2]float64{
	{131.0, 34.1}, {132.0, 35.0}, {133.0, 35.5}, {134.2, 35.7},
	{135.0, 35.5}, {135.0, 34.8}, {134.2, 34.3}, {133.2, 34.6},
	{132.0, 34.1}, {131.0, 34.1},
}

// Korean peninsula — home of Goguryeo, Baekje, and Silla during the Three Kingdoms period.
var koreaBorder = [][2]float64{
	{124.2, 40.0}, {125.0, 40.5}, {126.5, 41.8}, {129.5, 42.3},
	{130.5, 42.5}, {130.2, 41.0}, {129.8, 39.5}, {129.5, 38.5},
	{129.2, 37.5}, {128.8, 36.5}, {128.5, 35.5}, {128.0, 34.8},
	{127.0, 34.5}, {126.5, 35.0}, {126.0, 35.8}, {125.5, 36.5},
	{125.0, 37.5}, {124.5, 38.5}, {124.2, 39.5}, {124.2, 40.0},
}

// Approximate geographic coordinates for Three Kingdoms cities [lon, lat].
var cityGeo = map[string][2]float64{
	"Luoyang":    {112.5, 34.7}, "Xuchang":   {113.8, 34.0},
	"Chenliu":    {114.3, 34.8}, "Wancheng":  {112.5, 33.0},
	"Ye":         {114.6, 36.3}, "Youzhou":   {116.4, 40.0},
	"Qingzhou":   {118.5, 36.5}, "Bingzhou":  {112.5, 37.9},
	"Chang'an":   {108.9, 34.3}, "Tongguan":  {110.3, 34.5},
	"Tianshui":   {105.7, 34.6}, "Wudu":      {104.9, 33.4},
	"Xuzhou":     {117.2, 34.3}, "Xiapi":     {118.0, 34.0},
	"Hefei":      {117.3, 31.8}, "Xinye":     {112.4, 32.5},
	"Xiangyang":  {112.1, 32.0}, "Jiangling": {112.2, 30.3},
	"Changsha":   {113.0, 28.2}, "Lingling":  {111.6, 26.4},
	"Hanzhong":   {107.0, 33.1}, "Jiamenguan": {105.5, 32.0},
	"Chengdu":    {104.1, 30.7}, "Jiangzhou": {106.5, 29.6},
	"Yong'an":    {109.5, 31.0}, "Chaisang":  {116.0, 29.7},
	"Wuchang":    {114.9, 30.4}, "Jianye":    {118.8, 32.1},
	"Kuaiji":     {120.6, 30.0}, "Poyang":    {116.7, 29.0},
	// Nanzhong
	"Yuesui":     {102.2, 27.9}, "Jianning":  {103.8, 25.5},
	"Yongchang":  {99.2, 25.1},
	// Jiao Province
	"Panyu":      {113.2, 23.1}, "Jiaozhi":   {105.8, 21.0},
	// Korea
	"Lelang":       {126.0, 38.9},
	"Gungnaeseong": {126.5, 41.1},
	"Saro":          {129.2, 35.8},
	// Wa (Japan)
	"Ito":           {130.2, 33.5},
	"Yamatai":       {130.6, 33.1},
	"Izumo":         {132.7, 35.4},
	// Vietnam
	"Jiuzhen":       {105.5, 19.5},
}

// Mountain ranges [lonMin, latMin, lonMax, latMax] — rendered at checkerboard density.
var mountainRanges = [][4]float64{
	{103.5, 33.0, 112.0, 35.2}, // Qinling (秦岭) — the great Wei/Shu divide
	{112.0, 35.0, 114.2, 40.5}, // Taihang (太行) — eastern highland wall
	{105.0, 31.0, 110.5, 33.2}, // Daba/Micang (大巴/米仓) — Shu's north frontier
	{101.0, 29.0, 105.5, 33.5}, // Min Shan (岷山) — western Shu barrier
	{109.0, 40.0, 121.0, 42.5}, // Yanshan (燕山) — Wei's northern frontier
	{77.0, 27.0, 97.0, 30.0},   // Himalayas (喜马拉雅) — southern Tibet wall
	{74.0, 35.0, 100.0, 38.0},  // Kunlun (昆仑山) — central axis
	{73.0, 40.5, 94.0, 44.0},   // Tianshan (天山) — Xinjiang spine
	{87.0, 45.5, 100.0, 50.5},  // Mongolian Altai (阿尔泰山)
	{97.0, 24.0, 102.0, 30.0},  // Hengduan (横断山) — Yunnan-Sichuan SW
	{110.0, 23.5, 118.0, 26.5}, // Nanling (南岭) — south China divide
	{117.0, 25.0, 120.5, 29.0}, // Wuyi (武夷山) — SE coastal ranges
	{119.5, 40.0, 127.0, 42.5}, // Da Hinggan spur (大兴安岭 south)
}

// Forest regions [lonMin, latMin, lonMax, latMax] — sparse organic texture inside chinaBorder.
var forestRegions = [][4]float64{
	{120.0, 40.0, 135.0, 53.5}, // Manchurian taiga (大兴安岭/小兴安岭)
	{98.0, 24.5, 106.0, 30.0},  // Yunnan subtropical
	{106.0, 22.0, 112.0, 27.0}, // Guizhou/Guangxi
	{113.0, 22.5, 119.0, 27.0}, // South China (Nanling forests)
	{115.0, 27.0, 121.0, 31.5}, // Fujian/Jiangxi
	{107.0, 28.0, 114.0, 32.0}, // Central south Hunan
	{102.0, 30.0, 108.0, 34.0}, // Sichuan basin edge / Qinba fringe
}

// Marsh regions [lonMin, latMin, lonMax, latMax] — wetland and lake basins.
var marshRegions = [][4]float64{
	{115.0, 28.5, 122.0, 32.5}, // Jiangnan / Yangtze delta
	{112.0, 28.0, 116.0, 30.5}, // Dongting Lake (洞庭湖)
	{116.0, 28.5, 120.0, 31.0}, // Poyang Lake (鄱阳湖) basin
}

// Hills regions [lonMin, latMin, lonMax, latMax] — rolling terrain between plains and mountains.
var hillsRegions = [][4]float64{
	{106.0, 27.0, 113.0, 31.0}, // Central south (Guizhou/Hunan border)
	{113.0, 30.5, 118.5, 33.5}, // Dabie Shan foothills (大别山)
	{118.0, 32.0, 122.5, 35.0}, // Jiangsu/Zhejiang coastal hills
}

// Steppe regions [lonMin, latMin, lonMax, latMax] — open grassland north of the Great Wall.
var steppeRegions = [][4]float64{
	{100.0, 42.0, 122.0, 50.5}, // Mongolian plateau
	{82.0, 40.0, 100.0, 48.0},  // Central Asian steppe
	{119.0, 38.5, 128.0, 42.5}, // Manchurian-Mongolian fringe
}

// Rivers as polyline waypoints [lon, lat] — drawn as continuous lines.
var rivers = [][][2]float64{
	// Yellow River (黄河) — the great northern bend
	{{96.0, 35.5}, {100.0, 36.5}, {104.5, 37.5}, {107.0, 40.8},
		{109.0, 40.3}, {110.0, 38.5}, {110.5, 35.5}, {111.5, 35.0},
		{113.5, 35.0}, {115.5, 35.5}, {117.0, 36.3}, {118.5, 37.5}, {119.5, 38.5}},
	// Yangtze River (长江) — Sichuan to East China Sea
	{{99.0, 30.0}, {102.5, 29.2}, {104.6, 28.8}, {106.0, 29.0},
		{108.5, 30.2}, {110.5, 30.5}, {112.0, 30.0}, {113.5, 30.4},
		{114.9, 30.5}, {116.5, 30.0}, {118.0, 30.2}, {119.5, 31.5}, {121.5, 31.5}},
	// Han River (汉水) — Hanzhong to Xiangyang
	{{107.0, 33.1}, {109.0, 32.5}, {110.5, 32.3}, {112.1, 32.0}},
	// Wei River (渭河) — Gansu to Yellow River
	{{104.5, 34.4}, {106.0, 34.2}, {108.0, 34.3}, {110.0, 34.5}},
	// Liao River (辽河) — NE Manchuria
	{{119.5, 42.5}, {121.0, 41.5}, {122.0, 40.8}},
	// Pearl River (珠江) — south China
	{{110.0, 23.5}, {111.5, 23.3}, {112.5, 23.1}, {113.5, 23.1}},
}

// Sea regions [lonMin, latMin, lonMax, latMax] — land pixels excluded via pointInPoly.
var seaRegions = [][4]float64{
	{119.5, 18.0, 136.0, 40.5}, // East China Sea + Yellow Sea + western Pacific south
	{118.0, 38.0, 126.0, 42.5}, // Bohai Sea
	{127.5, 32.0, 136.0, 50.0}, // Sea of Japan + north approaches
	{110.0, 18.0, 122.5, 25.0}, // South China Sea (east of Vietnam coast)
	{107.0, 18.0, 111.0, 21.5}, // Gulf of Tonkin / Beibu Gulf (北部湾)
}

type canvas struct{ dots []bool }

func newCanvas() *canvas { return &canvas{dots: make([]bool, pw*ph)} }

func (c *canvas) set(x, y int) {
	if x >= 0 && x < pw && y >= 0 && y < ph {
		c.dots[y*pw+x] = true
	}
}

func (c *canvas) line(x0, y0, x1, y1 int) {
	dx, dy := iabs(x1-x0), iabs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		c.set(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (c *canvas) render() string {
	var sb strings.Builder
	for row := 0; row < mapH; row++ {
		for col := 0; col < mapW; col++ {
			px, py := col*2, row*4
			var b rune
			if c.dots[(py+0)*pw+(px+0)] {
				b |= 0x01
			}
			if c.dots[(py+1)*pw+(px+0)] {
				b |= 0x02
			}
			if c.dots[(py+2)*pw+(px+0)] {
				b |= 0x04
			}
			if c.dots[(py+3)*pw+(px+0)] {
				b |= 0x40
			}
			if c.dots[(py+0)*pw+(px+1)] {
				b |= 0x08
			}
			if c.dots[(py+1)*pw+(px+1)] {
				b |= 0x10
			}
			if c.dots[(py+2)*pw+(px+1)] {
				b |= 0x20
			}
			if c.dots[(py+3)*pw+(px+1)] {
				b |= 0x80
			}
			sb.WriteRune(0x2800 + b)
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// geoToPixel converts geographic coordinates to braille pixel coordinates.
func geoToPixel(lon, lat float64) (int, int) {
	x := int((lon - lonMin) / ((lonMax - lonMin) / xStretch) * float64(pw-1))
	y := int((latMax - lat) / (latMax - latMin) * float64(ph-1))
	return x, y
}

func pixelToGeo(px, py int) (float64, float64) {
	lon := float64(px)/float64(pw-1)*((lonMax-lonMin)/xStretch) + lonMin
	lat := latMax - float64(py)/float64(ph-1)*(latMax-latMin)
	return lon, lat
}

// pointInPoly uses ray casting to test whether (lon, lat) is inside a polygon.
func pointInPoly(lon, lat float64, poly [][2]float64) bool {
	inside := false
	j := len(poly) - 1
	for i := 0; i < len(poly); i++ {
		xi, yi := poly[i][0], poly[i][1]
		xj, yj := poly[j][0], poly[j][1]
		if ((yi > lat) != (yj > lat)) && lon < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}

func drawPoly(c *canvas, poly [][2]float64) {
	for i := 1; i < len(poly); i++ {
		x0, y0 := geoToPixel(poly[i-1][0], poly[i-1][1])
		x1, y1 := geoToPixel(poly[i][0], poly[i][1])
		c.line(x0, y0, x1, y1)
	}
}

// RenderMap returns a coloured braille string with layered terrain:
//
//	selected city (magenta) > cities (bold gold) > border (seasonal) > mountains > hills > rivers > marsh > forests > steppe > sea
func RenderMap(cities []models.City, cityPhase int, season string, wavePhase int, selectedCity string) string {
	// --- borders ---
	borders := newCanvas()
	drawPoly(borders, chinaBorder)
	drawPoly(borders, taiwanBorder)
	drawPoly(borders, kyushuBorder)
	drawPoly(borders, shikokuBorder)
	drawPoly(borders, honshuBorder)
	drawPoly(borders, koreaBorder)

	// --- city POIs ---
	citydots := newCanvas()
	for _, city := range cities {
		if geo, ok := cityGeo[city.Name]; ok {
			cx, cy := geoToPixel(geo[0], geo[1])
			citydots.set(cx, cy)
			citydots.set(cx+1, cy)
			citydots.set(cx, cy+1)
			citydots.set(cx+1, cy+1)
		}
	}

	// --- selected city highlight (same size as city dot, magenta) ---
	selecteddots := newCanvas()
	if selectedCity != "" {
		if geo, ok := cityGeo[selectedCity]; ok {
			cx, cy := geoToPixel(geo[0], geo[1])
			selecteddots.set(cx, cy)
			selecteddots.set(cx+1, cy)
			selecteddots.set(cx, cy+1)
			selecteddots.set(cx+1, cy+1)
		}
	}

	// --- mountain terrain (checkerboard, inside chinaBorder) ---
	mountains := newCanvas()
	for _, r := range mountainRanges {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if !pointInPoly(lon, lat, chinaBorder) {
					continue
				}
				if (px+py)%2 == 0 {
					mountains.set(px, py)
				}
			}
		}
	}

	// --- forest terrain (sparse organic scatter, inside chinaBorder) ---
	forests := newCanvas()
	for _, r := range forestRegions {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if !pointInPoly(lon, lat, chinaBorder) {
					continue
				}
				if (px*3+py*2)%7 == 0 {
					forests.set(px, py)
				}
			}
		}
	}

	// --- hills terrain (light diagonal pattern, inside chinaBorder) ---
	hillsCanvas := newCanvas()
	for _, r := range hillsRegions {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if !pointInPoly(lon, lat, chinaBorder) {
					continue
				}
				if (px+py)%5 == 0 {
					hillsCanvas.set(px, py)
				}
			}
		}
	}

	// --- marsh terrain (medium dot pattern, inside chinaBorder) ---
	marshCanvas := newCanvas()
	for _, r := range marshRegions {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if !pointInPoly(lon, lat, chinaBorder) {
					continue
				}
				if (px*2+py)%6 == 0 {
					marshCanvas.set(px, py)
				}
			}
		}
	}

	// --- steppe terrain (very sparse scatter, inside chinaBorder northern regions) ---
	steppeCanvas := newCanvas()
	for _, r := range steppeRegions {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if !pointInPoly(lon, lat, chinaBorder) {
					continue
				}
				if (px*4+py*3)%11 == 0 {
					steppeCanvas.set(px, py)
				}
			}
		}
	}

	// --- rivers (continuous polylines) ---
	riverCanvas := newCanvas()
	for _, river := range rivers {
		for i := 1; i < len(river); i++ {
			x0, y0 := geoToPixel(river[i-1][0], river[i-1][1])
			x1, y1 := geoToPixel(river[i][0], river[i][1])
			riverCanvas.line(x0, y0, x1, y1)
		}
	}

	// --- sea (animated diagonal ripple, masked by chinaBorder + 1px coastline buffer) ---
	water := newCanvas()
	for _, r := range seaRegions {
		x0, y1 := geoToPixel(r[0], r[1])
		x1, y0 := geoToPixel(r[2], r[3])
		for py := y0; py <= y1; py++ {
			for px := x0; px <= x1; px++ {
				lon, lat := pixelToGeo(px, py)
				if pointInPoly(lon, lat, chinaBorder) || pointInPoly(lon, lat, koreaBorder) ||
					pointInPoly(lon, lat, taiwanBorder) || pointInPoly(lon, lat, kyushuBorder) ||
					pointInPoly(lon, lat, shikokuBorder) || pointInPoly(lon, lat, honshuBorder) {
					continue
				}
				if (px+py*2+wavePhase)%8 < 2 {
					water.set(px, py)
				}
			}
		}
	}
	// erode water 1 pixel away from coastline for visual separation
	for py := 0; py < ph; py++ {
		for px := 0; px < pw; px++ {
			if !water.dots[py*pw+px] {
				continue
			}
			if (px > 0 && borders.dots[py*pw+(px-1)]) ||
				(px < pw-1 && borders.dots[py*pw+(px+1)]) ||
				(py > 0 && borders.dots[(py-1)*pw+px]) ||
				(py < ph-1 && borders.dots[(py+1)*pw+px]) {
				water.dots[py*pw+px] = false
			}
		}
	}

	// --- styles ---
	borderColour := lipgloss.Color("28")
	if c, ok := seasonBorderColour[season]; ok {
		borderColour = c
	}
	styleBorder := lipgloss.NewStyle().Bold(true).Foreground(borderColour)

	// ping-pong 6 gradient steps: 0→5→0
	step := cityPhase % 10
	if step >= 5 {
		step = 10 - step - 1
	}
	styleCity := mapStyleCity[step]

	// --- merge layers: selected > cities > border > mountains > hills > rivers > marsh > forests > steppe > sea ---
	borderRunes   := []rune(borders.render())
	selectedRunes := []rune(selecteddots.render())
	cityRunes     := []rune(citydots.render())
	mountainRunes := []rune(mountains.render())
	hillsRunes    := []rune(hillsCanvas.render())
	riverRunes    := []rune(riverCanvas.render())
	marshRunes    := []rune(marshCanvas.render())
	forestRunes   := []rune(forests.render())
	steppeRunes   := []rune(steppeCanvas.render())
	waterRunes    := []rune(water.render())

	var sb strings.Builder
	for i, br := range borderRunes {
		xr  := selectedRunes[i]
		cr  := cityRunes[i]
		mr  := mountainRunes[i]
		hr  := hillsRunes[i]
		rr  := riverRunes[i]
		mar := marshRunes[i]
		fr  := forestRunes[i]
		sr  := steppeRunes[i]
		wr  := waterRunes[i]
		switch {
		case br == '\n':
			sb.WriteRune('\n')
		case xr != 0x2800:
			sb.WriteString(mapStyleSelected.Render(string(xr)))
		case cr != 0x2800:
			sb.WriteString(styleCity.Render(string(cr)))
		case br != 0x2800:
			sb.WriteString(styleBorder.Render(string(br)))
		case mr != 0x2800:
			sb.WriteString(mapStyleMountain.Render(string(mr)))
		case hr != 0x2800:
			sb.WriteString(mapStyleHills.Render(string(hr)))
		case rr != 0x2800:
			sb.WriteString(mapStyleRiver.Render(string(rr)))
		case mar != 0x2800:
			sb.WriteString(mapStyleMarsh.Render(string(mar)))
		case fr != 0x2800:
			sb.WriteString(mapStyleForest.Render(string(fr)))
		case sr != 0x2800:
			sb.WriteString(mapStyleSteppe.Render(string(sr)))
		case wr != 0x2800:
			sb.WriteString(mapStyleWave.Render(string(wr)))
		default:
			sb.WriteRune(0x2800)
		}
	}
	frame := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	return frame.Render(sb.String())
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
