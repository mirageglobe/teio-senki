class_name Ledger
extends RefCounted

var officers: Dictionary = {}
var cities: Dictionary = {}
var game_clock: Dictionary = {"year": 189, "month": 1}
var resources: Dictionary = {"grain": 0, "gold": 0}
var logs: Array[Dictionary] = []

var _snapshot: Dictionary = {}

func save_snapshot():
	_snapshot = {
		"officers": officers.duplicate(true),
		"cities": cities.duplicate(true),
		"game_clock": game_clock.duplicate()
	}

func restore_snapshot():
	officers = _snapshot.officers.duplicate(true)
	cities = _snapshot.cities.duplicate(true)
	game_clock = _snapshot.game_clock.duplicate()

func load_data():
	officers = _load_json("res://data/officers.json")
	cities = _load_json("res://data/cities.json")
	print("Ledger: Loaded %d officers and %d cities." % [officers.size(), cities.size()])

func _load_json(path: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		print("Error: File not found: ", path)
		return {}
	
	var file: FileAccess = FileAccess.open(path, FileAccess.READ)
	var json_text: String = file.get_as_text()
	var json: JSON = JSON.new()
	var error: int = json.parse(json_text)

	if error != OK:
		print("Error parsing JSON: ", json.get_error_message())
		return {}

	var data: Dictionary = json.get_data()
	var collection_key: String = data.keys()[0]
	var list: Array = data[collection_key]
	var result: Dictionary = {}

	for item: Dictionary in list:
		var id: String = item.get("name", "unknown")
		result[id] = item
		
	return result

func get_officer(id: String) -> Dictionary:
	return officers.get(id, {})

func get_city(id: String) -> Dictionary:
	return cities.get(id, {})

func log_event(type: String, description: String, effects: Dictionary = {}) -> void:
	var entry: Dictionary = {
		"year": game_clock.year,
		"month": game_clock.month,
		"type": type,
		"description": description,
		"effects": effects
	}
	logs.append(entry)
	print("[%d.%d] %s: %s" % [entry.year, entry.month, entry.type, entry.description])
