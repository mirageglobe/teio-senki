class_name TestCityEconomy
extends RefCounted

# TDD: these tests define the CityEconomy contract before implementation.
# Run via test_runner.gd.

static func run() -> bool:
	var success := true
	const CityEconomyScript = preload("res://scripts/core/city_economy.gd")

	# Test 1: base grain yield scales with agriculture; technology = 0 means no bonus
	var grain := CityEconomyScript.calc_grain_yield(80, 0)
	if grain == 80:
		print("[PASS] Grain yield: base agriculture, no tech bonus")
	else:
		print("[FAIL] Grain yield: expected 80, got %d" % grain)
		success = false

	# Test 2: technology amplifies grain yield (Fire generates Wood in Wu Xing)
	var grain_tech := CityEconomyScript.calc_grain_yield(80, 100)
	if grain_tech > 80:
		print("[PASS] Grain yield: technology bonus increases output")
	else:
		print("[FAIL] Grain yield: technology bonus had no effect")
		success = false

	# Test 3: base gold yield scales with commerce; technology = 0 means no bonus
	var gold := CityEconomyScript.calc_gold_yield(85, 0)
	if gold == 85:
		print("[PASS] Gold yield: base commerce, no tech bonus")
	else:
		print("[FAIL] Gold yield: expected 85, got %d" % gold)
		success = false

	# Test 4: technology amplifies gold yield
	var gold_tech := CityEconomyScript.calc_gold_yield(85, 100)
	if gold_tech > 85:
		print("[PASS] Gold yield: technology bonus increases output")
	else:
		print("[FAIL] Gold yield: technology bonus had no effect")
		success = false

	# Test 5: perfect order means zero corruption loss
	var loss_high_order := CityEconomyScript.calc_corruption_loss(100)
	if loss_high_order == 0:
		print("[PASS] Corruption: full order yields zero loss")
	else:
		print("[FAIL] Corruption: expected 0 loss at order=100, got %d" % loss_high_order)
		success = false

	# Test 6: zero order means maximum corruption loss
	var loss_no_order := CityEconomyScript.calc_corruption_loss(0)
	if loss_no_order > 0:
		print("[PASS] Corruption: zero order yields positive loss")
	else:
		print("[FAIL] Corruption: expected positive loss at order=0")
		success = false

	# Test 7: yields are clamped — no negative output
	var grain_zero := CityEconomyScript.calc_grain_yield(0, 0)
	if grain_zero >= 0:
		print("[PASS] Grain yield: non-negative floor holds")
	else:
		print("[FAIL] Grain yield: returned negative value")
		success = false

	return success
