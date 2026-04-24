extends Control

const LedgerScript = preload("res://scripts/engine/ledger.gd")

func _unhandled_input(event: InputEvent) -> void:
	var camera = $Camera2D
	if event is InputEventMouseButton:
		if event.button_index == MOUSE_BUTTON_WHEEL_UP:
			camera.zoom *= 1.1
		elif event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			camera.zoom /= 1.1
	elif event is InputEventKey:
		if event.pressed:
			if event.keycode == KEY_EQUAL: # + key
				camera.zoom *= 1.1
			elif event.keycode == KEY_MINUS: # - key
				camera.zoom /= 1.1

func _ready() -> void:
	var camera := Camera2D.new()
	camera.name = "Camera2D"
	camera.anchor_mode = Camera2D.ANCHOR_MODE_DRAG_CENTER
	camera.position = Vector2(640, 360)
	camera.zoom = Vector2(1.2, 1.2)
	add_child(camera)

	var ledger := LedgerScript.new()
	ledger.load_data()
	_render_cities(ledger.cities)

func _draw() -> void:
	draw_rect(Rect2(120, 100, 400, 350), Color(0.15, 0.2, 0.15), true) # mainland
	draw_rect(Rect2(520, 100, 80, 350), Color(0.1, 0.15, 0.1), true)   # coastal/sea

func _render_cities(cities: Dictionary) -> void:
	for city_name: String in cities:
		var city: Dictionary = cities[city_name]
		var btn := Button.new()
		btn.text = "● " + city.get("name", "Unknown")
		btn.add_theme_font_size_override("font_size", 10)
		btn.position = Vector2(city.x * 32, city.y * 32)
		btn.flat = true
		add_child(btn)

func _input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel"):
		get_tree().quit()
