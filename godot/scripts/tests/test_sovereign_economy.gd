class_name TestSovereignEconomy
extends RefCounted

# TDD: defines the contract for economic commands and per-turn yield before implementation.

static func run() -> bool:
	var success := true

	const LedgerScript = preload("res://scripts/engine/ledger.gd")
	const GameClockScript = preload("res://scripts/core/game_clock.gd")
	const EngineScript = preload("res://scripts/engine/sovereign_engine.gd")

	var ledger := LedgerScript.new()
	ledger.load_data()
	var clock := GameClockScript.new(189, 1)
	var engine := EngineScript.new(ledger, clock)

	# Test 1: BUILD_AG increases city agriculture stat
	var ag_before: int = ledger.get_city("Luoyang").get("agriculture", 0)
	engine.start_turn()
	engine.queue_command("BUILD_AG", {"city": "Luoyang"}, 2)
	engine.settle_turn()
	var ag_after: int = ledger.get_city("Luoyang").get("agriculture", 0)
	if ag_after > ag_before:
		print("[PASS] BUILD_AG: agriculture stat increased")
	else:
		print("[FAIL] BUILD_AG: expected agriculture to increase, was %d → %d" % [ag_before, ag_after])
		success = false

	# Test 2: BUILD_COM increases city commerce stat
	var com_before: int = ledger.get_city("Luoyang").get("commerce", 0)
	engine.start_turn()
	engine.queue_command("BUILD_COM", {"city": "Luoyang"}, 2)
	engine.settle_turn()
	var com_after: int = ledger.get_city("Luoyang").get("commerce", 0)
	if com_after > com_before:
		print("[PASS] BUILD_COM: commerce stat increased")
	else:
		print("[FAIL] BUILD_COM: expected commerce to increase, was %d → %d" % [com_before, com_after])
		success = false

	# Test 3: per-turn grain is accumulated in ledger.resources after settlement
	if ledger.resources.get("grain", 0) > 0:
		print("[PASS] Resources: grain accumulated after turns")
	else:
		print("[FAIL] Resources: no grain in ledger after turns")
		success = false

	# Test 4: per-turn gold is accumulated in ledger.resources after settlement
	if ledger.resources.get("gold", 0) > 0:
		print("[PASS] Resources: gold accumulated after turns")
	else:
		print("[FAIL] Resources: no gold in ledger after turns")
		success = false

	# Test 5: city agriculture is capped at 100
	for i in range(20):
		engine.start_turn()
		engine.queue_command("BUILD_AG", {"city": "Luoyang"}, 2)
		engine.settle_turn()
	var ag_capped: int = ledger.get_city("Luoyang").get("agriculture", 0)
	if ag_capped <= 100:
		print("[PASS] BUILD_AG: agriculture stat capped at 100")
	else:
		print("[FAIL] BUILD_AG: agriculture exceeded 100 (%d)" % ag_capped)
		success = false

	return success
