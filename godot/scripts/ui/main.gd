extends Control

const LedgerScript = preload("res://scripts/engine/ledger.gd")
const CityPanelDataScript = preload("res://scripts/core/city_panel_data.gd")

var _ledger: RefCounted
var _panel: PanelContainer
var _panel_label: RichTextLabel
var _ui_layer: CanvasLayer

func _ready() -> void:
	var camera := Camera2D.new()
	camera.name = "Camera2D"
	camera.anchor_mode = Camera2D.ANCHOR_MODE_DRAG_CENTER
	camera.position = Vector2(640, 360)
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

func _draw() -> void:
	draw_rect(Rect2(120, 100, 400, 350), Color(0.15, 0.2, 0.15), true)
	draw_rect(Rect2(520, 100, 80, 350), Color(0.1, 0.15, 0.1), true)

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
		container.position = Vector2(city.x * 32 - 32, city.y * 32 - 8)
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

