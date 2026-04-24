class_name Ledger
extends RefCounted

# The Ledger manages the game state.
# Initially using JSON files; will transition to SQLite for production.

var officers: Dictionary = {}
var cities: Dictionary = {}
var game_clock: Dictionary = {"year": 189, "month": 1}
var logs: Array = []

func load_data():
	officers = _load_json("res://data/officers.json")
	cities = _load_json("res://data/cities.json")
	print("Ledger: Loaded %d officers and %d cities." % [officers.size(), cities.size()])

func _load_json(path: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		print("Error: File not found: ", path)
		return {}
	
	var file = FileAccess.open(path, FileAccess.READ)
	var json_text = file.get_as_text()
	var json = JSON.new()
	var error = json.parse(json_text)
	
	if error != OK:
		print("Error parsing JSON: ", json.get_error_message())
		return {}
		
	var data = json.get_data()
	
	# If the root is a list (like our YAML structure), convert to a dict by ID
	# We'll assume the first key is the collection name (e.g., "officers")
	var collection_key = data.keys()[0]
	var list = data[collection_key]
	var result = {}
	
	for item in list:
		var id = item.get("name", "unknown") # Use name as ID for now
		result[id] = item
		
	return result

func get_officer(id: String) -> Dictionary:
	return officers.get(id, {})

func get_city(id: String) -> Dictionary:
	return cities.get(id, {})

func log_event(type: String, description: String, effects: Dictionary = {}):
	var entry = {
		"year": game_clock.year,
		"month": game_clock.month,
		"type": type,
		"description": description,
		"effects": effects
	}
	logs.append(entry)
	print("[%d.%d] %s: %s" % [entry.year, entry.month, entry.type, entry.description])
