extends Control

const LedgerScript = preload("res://scripts/engine/ledger.gd")
const CityPanelDataScript = preload("res://scripts/core/city_panel_data.gd")

var _ledger: RefCounted
var _panel: PanelContainer
var _panel_label: RichTextLabel
var _ui_layer: CanvasLayer

func _ready() -> void:
	RenderingServer.set_default_clear_color(Color(0.04, 0.04, 0.04, 1))
	
	var camera := Camera2D.new()
	camera.name = "Camera2D"
	camera.anchor_mode = Camera2D.ANCHOR_MODE_DRAG_CENTER
	camera.position = Vector2(320, 240)
	camera.zoom = Vector2(1.2, 1.2)
	add_child(camera)

	# ui layer sits above Camera2D transforms — always screen-space
	_ui_layer = CanvasLayer.new()
	_ui_layer.layer = 10
	add_child(_ui_layer)

	_ledger = LedgerScript.new()
	_ledger.load_data()

	_render_cities(_ledger.cities)
	_build_city_panel()

func _grid_to_pixel(x: float, y: float) -> Vector2:
	# Y is inverted: y=0 is South (bottom of map), y=14 is North (top of map)
	return Vector2(x * 32.0, (14.0 - y) * 32.0)

func _smooth_points(points: Array, subdivisions: int, closed: bool = false) -> PackedVector2Array:
	var result := PackedVector2Array()
	var n := points.size()
	if n < 2: return result
	
	for i in range(n):
		if not closed and i == n - 1:
			result.append(_grid_to_pixel(points[i].x, points[i].y))
			break
			
		var p_prev: Vector2
		var p_curr: Vector2 = points[i]
		var p_next: Vector2
		var p_next2: Vector2
		
		if closed:
			p_prev = points[posmod(i - 1, n)]
			p_next = points[posmod(i + 1, n)]
			p_next2 = points[posmod(i + 2, n)]
		else:
			p_prev = points[0] if i == 0 else points[i - 1]
			p_next = points[i + 1]
			p_next2 = points[n - 1] if i == n - 2 else points[i + 2]
		
		var v_prev := _grid_to_pixel(p_prev.x, p_prev.y)
		var v_curr := _grid_to_pixel(p_curr.x, p_curr.y)
		var v_next := _grid_to_pixel(p_next.x, p_next.y)
		var v_next2 := _grid_to_pixel(p_next2.x, p_next2.y)
		
		for j in range(subdivisions):
			var t := float(j) / float(subdivisions)
			var pt := v_curr.cubic_interpolate(v_next, v_prev, v_next2, t)
			result.append(pt)
			
	return result

func _draw() -> void:
	var map_points := [
		# Hexi Corridor
		Vector2(0.5, 11.5), Vector2(1.5, 12.2), Vector2(2.5, 12.8),
		# Northern Steppes Border (Inner Mongolia)
		Vector2(4.0, 13.1), Vector2(6.0, 13.5), Vector2(8.0, 13.8), Vector2(10.0, 14.1), Vector2(12.0, 14.3),
		# Northeast into Liaodong
		Vector2(13.5, 14.5), Vector2(15.0, 14.8), Vector2(16.5, 14.6), Vector2(17.5, 13.8), Vector2(17.0, 12.8), Vector2(16.2, 12.5),
		# Bohai Sea (Inward curve)
		Vector2(15.2, 12.6), Vector2(14.5, 12.9), Vector2(13.8, 12.7), Vector2(13.5, 12.2),
		# Shandong Peninsula (Bulges East)
		Vector2(14.2, 11.8), Vector2(15.2, 11.7), Vector2(16.5, 11.5), Vector2(17.0, 10.8), Vector2(16.5, 10.0), Vector2(15.5, 9.6),
		# Eastern Coastline (Jiangsu)
		Vector2(14.8, 8.8), Vector2(14.6, 8.0), Vector2(15.2, 7.2),
		# Jiangdong (Bulges East around Yangtze delta)
		Vector2(16.5, 6.8), Vector2(17.2, 6.0), Vector2(17.0, 5.0), Vector2(16.2, 4.0), Vector2(15.5, 3.2),
		# South Coast (Fujian/Guangdong/Guangxi)
		Vector2(14.8, 2.4), Vector2(14.0, 1.6), Vector2(13.0, 1.0), Vector2(11.8, 0.5), Vector2(10.2, 0.0), Vector2(8.5, -0.2), Vector2(6.5, -0.1),
		# Southwest (Vietnam/Yunnan border)
		Vector2(5.0, 0.4), Vector2(3.8, 1.2), Vector2(2.8, 2.2), Vector2(2.2, 3.0),
		# Western borders (Sichuan Basin / Tibetan Plateau)
		Vector2(1.8, 4.2), Vector2(1.6, 5.5), Vector2(2.0, 6.8), Vector2(2.8, 8.0), Vector2(3.2, 8.8),
		# Up towards Gansu
		Vector2(2.5, 9.8), Vector2(1.5, 10.5)
	]
	
	var hainan_points := [
		Vector2(8.5, -0.8), Vector2(9.2, -0.6), Vector2(9.5, -1.2), Vector2(8.8, -1.6), Vector2(8.2, -1.2)
	]
	
	var taiwan_points := [
		Vector2(16.5, 2.5), Vector2(17.2, 2.8), Vector2(17.0, 1.5), Vector2(16.2, 1.2)
	]
	
	var land_color := Color(0.12, 0.14, 0.12)
	var border_color := Color(0.2, 0.25, 0.2)
	
	# Draw Mainland
	var poly := _smooth_points(map_points, 6, true)
	draw_colored_polygon(poly, land_color)
	var main_border := poly
	main_border.append(poly[0])
	draw_polyline(main_border, border_color, 2.0, true)
	
	# Draw Hainan
	var hainan_poly := _smooth_points(hainan_points, 4, true)
	draw_colored_polygon(hainan_poly, land_color)
	var hainan_border := hainan_poly
	hainan_border.append(hainan_poly[0])
	draw_polyline(hainan_border, border_color, 1.5, true)
	
	# Draw Taiwan
	var taiwan_poly := _smooth_points(taiwan_points, 4, true)
	draw_colored_polygon(taiwan_poly, land_color)
	var taiwan_border := taiwan_poly
	taiwan_border.append(taiwan_poly[0])
	draw_polyline(taiwan_border, border_color, 1.5, true)

	# Draw Rivers
	var huang_he_points := [
		Vector2(1.0, 9.0),   # Origin west
		Vector2(3.0, 9.5),
		Vector2(4.5, 10.5),  # Near Tianshui
		Vector2(5.2, 11.8),
		Vector2(5.5, 12.8),  # Up the Ordos Loop
		Vector2(6.5, 13.0),  # Top of Ordos
		Vector2(7.5, 12.8),  # Top of Ordos Loop
		Vector2(8.0, 11.5),
		Vector2(8.5, 10.5),  # Down towards Luoyang/Chang'an
		Vector2(9.5, 10.2),
		Vector2(11.0, 10.8), # Past Luoyang
		Vector2(12.5, 11.2),
		Vector2(13.8, 11.6), # Towards Bohai
		Vector2(14.8, 11.8)  # Out to sea
	]
	var huang_he := _smooth_points(huang_he_points, 6, false)
	draw_polyline(huang_he, Color(0.15, 0.25, 0.35, 0.8), 2.0, true)
	
	var chang_jiang_points := [
		Vector2(1.0, 3.5),   # Origin west of Shu
		Vector2(2.5, 4.0),
		Vector2(4.0, 4.5),   # Through Chengdu area
		Vector2(5.0, 4.2),
		Vector2(6.0, 4.5),   # Near Jiangzhou
		Vector2(7.5, 5.0),
		Vector2(8.5, 5.5),   # Jiangling
		Vector2(10.0, 5.2),
		Vector2(11.0, 5.5),  # Chaisang/Wuchang
		Vector2(12.5, 6.0),
		Vector2(13.5, 6.5),  # Jianye
		Vector2(15.0, 6.2),
		Vector2(16.5, 6.5)   # Out to sea
	]
	var chang_jiang := _smooth_points(chang_jiang_points, 6, false)
	draw_polyline(chang_jiang, Color(0.15, 0.25, 0.35, 0.8), 2.0, true)

func _render_cities(cities: Dictionary) -> void:
	for city_name: String in cities:
		var city: Dictionary = cities[city_name]

		var dot := Button.new()
		dot.text = "●"
		dot.add_theme_font_size_override("font_size", 10)
		dot.flat = true
		dot.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
		dot.pressed.connect(_on_city_pressed.bind(city_name))

		var lbl := Label.new()
		lbl.text = city.get("name", "Unknown")
		lbl.add_theme_font_size_override("font_size", 8)
		lbl.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		lbl.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
		lbl.visible = false

		dot.mouse_entered.connect(func() -> void: lbl.visible = true)
		dot.mouse_exited.connect(func() -> void: lbl.visible = false)

		var container := VBoxContainer.new()
		# centre the 64px-wide container on the grid point
		var center := _grid_to_pixel(float(city.x), float(city.y))
		container.position = center + Vector2(-32, -12)
		container.custom_minimum_size = Vector2(64, 0)
		container.add_child(dot)
		container.add_child(lbl)
		add_child(container)

func _build_city_panel() -> void:
	_panel = PanelContainer.new()
	_panel.anchor_left = 1.0
	_panel.anchor_right = 1.0
	_panel.anchor_top = 0.0
	_panel.anchor_bottom = 1.0
	_panel.offset_left = -240
	_panel.offset_right = 0
	_panel.offset_top = 0
	_panel.offset_bottom = 0
	_panel.visible = false

	var margin := MarginContainer.new()
	margin.size_flags_vertical = Control.SIZE_EXPAND_FILL
	margin.add_theme_constant_override("margin_top", 12)
	margin.add_theme_constant_override("margin_right", 10)
	margin.add_theme_constant_override("margin_bottom", 12)
	margin.add_theme_constant_override("margin_left", 10)

	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED

	_panel_label = RichTextLabel.new()
	_panel_label.bbcode_enabled = true
	_panel_label.fit_content = true
	_panel_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_panel_label.add_theme_font_size_override("normal_font_size", 11)

	scroll.add_child(_panel_label)
	margin.add_child(scroll)
	_panel.add_child(margin)
	_ui_layer.add_child(_panel)

func _on_city_pressed(city_name: String) -> void:
	var city: Dictionary = _ledger.get_city(city_name)
	var data: Dictionary = CityPanelDataScript.format(city)

	var pillars: Dictionary = data.get("pillars", {})
	_panel_label.text = (
		"[b]%s[/b]\n" % data.get("title", city_name) +
		"[color=#888888]%s  ·  %s[/color]\n" % [data.get("region", ""), data.get("terrain", "")] +
		"[color=#888888]%s[/color]\n" % data.get("faction", "—") +
		"Pop  %s\n" % _fmt_pop(data.get("population", 0)) +
		"\n[b]— pillars —[/b]\n" +
		"Agriculture   [b]%d[/b]\n" % pillars.get("agriculture", 0) +
		"Commerce      [b]%d[/b]\n" % pillars.get("commerce", 0) +
		"Technology    [b]%d[/b]\n" % pillars.get("technology", 0) +
		"Order         [b]%d[/b]\n" % pillars.get("order", 0) +
		"\n[b]— yield / turn —[/b]\n" +
		"Grain    [color=#88cc88]+%d[/color]\n" % data.get("grain_yield", 0) +
		"Gold     [color=#cccc66]+%d[/color]\n" % data.get("gold_yield", 0) +
		"Corrupt  [color=#cc6666]-%d[/color]\n" % data.get("corruption_loss", 0) +
		"\n[color=#555555][i]esc to close[/i][/color]"
	)
	_panel.visible = true

static func _fmt_pop(pop: int) -> String:
	if pop >= 1000000:
		return "%.1fm" % (pop / 1000000.0)
	if pop >= 1000:
		return "%dk" % (pop / 1000)
	return str(pop)

func _unhandled_input(event: InputEvent) -> void:
	var camera: Camera2D = $Camera2D
	if event is InputEventMouseButton:
		if event.button_index == MOUSE_BUTTON_WHEEL_UP:
			camera.zoom *= 1.1
		elif event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			camera.zoom /= 1.1
	elif event is InputEventKey and event.pressed:
		match event.keycode:
			KEY_EQUAL:
				camera.zoom *= 1.1
			KEY_MINUS:
				camera.zoom /= 1.1
			KEY_ESCAPE:
				if _panel.visible:
					_panel.visible = false
				else:
					get_tree().quit()
