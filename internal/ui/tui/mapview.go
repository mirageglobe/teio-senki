package tui

import (
	"strings"

	"github.com/mirageglobe/teio-senki/internal/models"
)

const mapW, mapH = 26, 11 // braille chars; pixel grid = 52×44
const pw, ph = mapW * 2, mapH * 4

const lonMin, lonMax = 73.0, 135.0
const latMin, latMax = 18.0, 53.0

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
	"Hanzhong":   {107.0, 33.1}, "Jiamenguan":{105.5, 32.0},
	"Chengdu":    {104.1, 30.7}, "Jiangzhou": {106.5, 29.6},
	"Yong'an":    {109.5, 31.0}, "Chaisang":  {116.0, 29.7},
	"Wuchang":    {114.9, 30.4}, "Jianye":    {118.8, 32.1},
	"Kuaiji":     {120.6, 30.0}, "Poyang":    {116.7, 29.0},
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
			if c.dots[(py+0)*pw+(px+0)] { b |= 0x01 }
			if c.dots[(py+1)*pw+(px+0)] { b |= 0x02 }
			if c.dots[(py+2)*pw+(px+0)] { b |= 0x04 }
			if c.dots[(py+3)*pw+(px+0)] { b |= 0x40 }
			if c.dots[(py+0)*pw+(px+1)] { b |= 0x08 }
			if c.dots[(py+1)*pw+(px+1)] { b |= 0x10 }
			if c.dots[(py+2)*pw+(px+1)] { b |= 0x20 }
			if c.dots[(py+3)*pw+(px+1)] { b |= 0x80 }
			sb.WriteRune(0x2800 + b)
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

func geoToPixel(lon, lat float64) (int, int) {
	x := int((lon - lonMin) / (lonMax - lonMin) * float64(pw-1))
	y := int((latMax - lat) / (latMax - latMin) * float64(ph-1))
	return x, y
}

// RenderMap returns a braille dot string: China outline + city dots.
func RenderMap(cities []models.City) string {
	c := newCanvas()
	for i := 1; i < len(chinaBorder); i++ {
		x0, y0 := geoToPixel(chinaBorder[i-1][0], chinaBorder[i-1][1])
		x1, y1 := geoToPixel(chinaBorder[i][0], chinaBorder[i][1])
		c.line(x0, y0, x1, y1)
	}
	for _, city := range cities {
		if geo, ok := cityGeo[city.Name]; ok {
			cx, cy := geoToPixel(geo[0], geo[1])
			c.set(cx, cy)
			c.set(cx+1, cy)
			c.set(cx, cy+1)
			c.set(cx+1, cy+1)
		}
	}
	return c.render()
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
