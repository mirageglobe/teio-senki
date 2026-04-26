class_name TestCityPanelData
extends RefCounted

static func run() -> bool:
	var success := true
	const CityPanelDataScript = preload("res://scripts/core/city_panel_data.gd")

	var city := {
		"name": "Luoyang", "chinese": "洛陽", "region": "Si Li",
		"terrain": "plain", "faction": "Dong Zhuo",
		"population": 500000, "defense": 80,
		"agriculture": 80, "commerce": 85, "technology": 80, "order": 35
	}

	var panel: Dictionary = CityPanelDataScript.format(city)

	# Test 1: panel includes display name with chinese characters
	if "洛陽" in panel.get("title", ""):
		print("[PASS] Panel: title includes chinese name")
	else:
		print("[FAIL] Panel: title missing chinese name — got: %s" % panel.get("title", ""))
		success = false

	# Test 2: panel includes faction
	if panel.get("faction", "") == "Dong Zhuo":
		print("[PASS] Panel: faction present")
	else:
		print("[FAIL] Panel: faction missing or wrong")
		success = false

	# Test 3: panel includes all four pillar stats
	for pillar in ["agriculture", "commerce", "technology", "order"]:
		if pillar in panel.get("pillars", {}):
			print("[PASS] Panel: %s pillar present" % pillar)
		else:
			print("[FAIL] Panel: %s pillar missing" % pillar)
			success = false

	# Test 4: panel includes computed grain and gold yields
	if panel.get("grain_yield", -1) > 0:
		print("[PASS] Panel: grain_yield computed")
	else:
		print("[FAIL] Panel: grain_yield missing or zero")
		success = false

	if panel.get("gold_yield", -1) > 0:
		print("[PASS] Panel: gold_yield computed")
	else:
		print("[FAIL] Panel: gold_yield missing or zero")
		success = false

	# Test 5: empty dict returns safe defaults (no crash)
	var empty_panel: Dictionary = CityPanelDataScript.format({})
	if empty_panel.get("title", "") != "":
		print("[PASS] Panel: empty city returns safe defaults")
	else:
		print("[FAIL] Panel: empty city caused missing title")
		success = false

	return success
