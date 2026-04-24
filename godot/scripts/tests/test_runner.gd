extends SceneTree

# Headless Test Runner
# Run via: godot --path godot/ --headless -s scripts/tests/test_runner.gd

func _init() -> void:
	print("--- BEGINNING HEADLESS ENGINE TESTS ---")
	
	# Load scripts dynamically
	var game_clock = load("res://scripts/core/game_clock.gd")
	var essence = load("res://scripts/core/essence.gd")
	var ledger = load("res://scripts/engine/ledger.gd")
	var engine = load("res://scripts/engine/sovereign_engine.gd")
	
	var success = true
	
	# Test 1: Bazi Clock
	var clock = game_clock.new(189, 12)
	clock.advance_month()
	if clock.year == 190 and clock.month == 1:
		print("[PASS] Clock progression")
	else:
		print("[FAIL] Clock progression")
		success = false

	# Test 2: Essence Drift
	if essence.get_drift_multiplier("Metal", "Metal") == 1.25:
		print("[PASS] Essence drift math")
	else:
		print("[FAIL] Essence drift math")
		success = false

	# Test 3: Ledger
	var l = ledger.new()
	l.load_data()
	if not l.get_officer("Cao Cao").is_empty():
		print("[PASS] Ledger data loading")
	else:
		print("[FAIL] Ledger loading")
		success = false

	# Test 4: Turn Loop
	var e = engine.new(l, clock)
	e.start_turn()
	e.queue_command("BUILD_AG", {"city": "Luoyang"}, 2)
	e.settle_turn()
	
	if not l.logs.is_empty():
		print("[PASS] Full turn loop")
	else:
		print("[FAIL] Turn loop")
		success = false

	# Test suite: City Economy
	var test_city_economy = load("res://scripts/tests/test_city_economy.gd")
	if not test_city_economy.run():
		success = false

	# Test suite: Sovereign Economy (build commands + per-turn yield)
	var test_sovereign_economy = load("res://scripts/tests/test_sovereign_economy.gd")
	if not test_sovereign_economy.run():
		success = false

	# Test suite: City Panel Data (formatter for UI)
	var test_city_panel_data = load("res://scripts/tests/test_city_panel_data.gd")
	if not test_city_panel_data.run():
		success = false

	print("--- TESTS COMPLETE ---")
	
	if success:
		print("STATUS: ALL TESTS PASSED")
		quit(0)
	else:
		print("STATUS: TESTS FAILED")
		quit(1)
