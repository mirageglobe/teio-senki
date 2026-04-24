extends Control

const LedgerScript = preload("res://scripts/engine/ledger.gd")

func _ready() -> void:
	var ledger = LedgerScript.new()
	ledger.load_data()
	_render_cities(ledger.cities)

func _draw():
	# Procedural stylized shapes for China regions
	# These are low-fi placeholders for the map topography
	draw_rect(Rect2(120, 100, 400, 350), Color(0.15, 0.2, 0.15), true) # Mainland
	draw_rect(Rect2(520, 100, 80, 350), Color(0.1, 0.15, 0.1), true)   # Coastal/Sea

func _render_cities(cities: Dictionary):
	for city_name in cities:
		var city = cities[city_name]
		var btn = Button.new()
		btn.text = "● " + city.get("name", "Unknown")
		btn.add_theme_font_size_override("font_size", 10)
		btn.position = Vector2(city.x * 32, city.y * 32)
		btn.flat = true
		add_child(btn)

func _input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel"):
		get_tree().quit()
