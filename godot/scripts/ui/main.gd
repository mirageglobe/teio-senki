extends Control

const LedgerScript = preload("res://scripts/engine/ledger.gd")
const CityPanelDataScript = preload("res://scripts/core/city_panel_data.gd")

var _ledger: RefCounted
var _panel: PanelContainer
var _panel_label: RichTextLabel

func _ready() -> void:
	var camera := Camera2D.new()
	camera.name = "Camera2D"
	camera.anchor_mode = Camera2D.ANCHOR_MODE_DRAG_CENTER
	camera.position = Vector2(640, 360)
	camera.zoom = Vector2(1.2, 1.2)
	add_child(camera)

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
		var btn := Button.new()
		btn.text = "● " + city.get("name", "Unknown")
		btn.add_theme_font_size_override("font_size", 10)
		btn.position = Vector2(city.x * 32, city.y * 32)
		btn.flat = true
		btn.pressed.connect(_on_city_pressed.bind(city_name))
		add_child(btn)

func _build_city_panel() -> void:
	_panel = PanelContainer.new()
	_panel.anchor_left = 1.0
	_panel.anchor_right = 1.0
	_panel.anchor_bottom = 1.0
	_panel.offset_left = -220
	_panel.offset_top = 0
	_panel.offset_right = 0
	_panel.offset_bottom = 0
	_panel.visible = false

	_panel_label = RichTextLabel.new()
	_panel_label.bbcode_enabled = true
	_panel_label.fit_content = true
	_panel_label.add_theme_font_size_override("normal_font_size", 11)
	_panel.add_child(_panel_label)

	add_child(_panel)

func _on_city_pressed(city_name: String) -> void:
	var city: Dictionary = _ledger.get_city(city_name)
	var data: Dictionary = CityPanelDataScript.format(city)

	var pillars: Dictionary = data.get("pillars", {})
	_panel_label.text = (
		"[b]%s[/b]\n" % data.get("title", city_name) +
		"%s · %s\n" % [data.get("region", ""), data.get("terrain", "")] +
		"Faction: %s\n" % data.get("faction", "—") +
		"Population: %s\n\n" % _fmt_pop(data.get("population", 0)) +
		"[b]Pillars[/b]\n" +
		"Agriculture  %d\n" % pillars.get("agriculture", 0) +
		"Commerce     %d\n" % pillars.get("commerce", 0) +
		"Technology   %d\n" % pillars.get("technology", 0) +
		"Order        %d\n\n" % pillars.get("order", 0) +
		"[b]Yield / Turn[/b]\n" +
		"Grain  +%d\n" % data.get("grain_yield", 0) +
		"Gold   +%d\n" % data.get("gold_yield", 0) +
		"Corruption  -%d\n" % data.get("corruption_loss", 0)
	)
	_panel.visible = true

static func _fmt_pop(pop: int) -> String:
	if pop >= 1000000:
		return "%.1fm" % (pop / 1000000.0)
	if pop >= 1000:
		return "%dk" % (pop / 1000)
	return str(pop)

func _unhandled_input(event: InputEvent) -> void:
	var camera = $Camera2D
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

func _input(event: InputEvent) -> void:
	pass
