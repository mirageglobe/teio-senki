class_name SovereignEngine
extends RefCounted

# Explicit preloads so class_name globals are not required — safe for headless -s mode.
const EssenceScript = preload("res://scripts/core/essence.gd")
const GameClockScript = preload("res://scripts/core/game_clock.gd")
const CityEconomyScript = preload("res://scripts/core/city_economy.gd")

const BUILD_GAIN := 5  # stat points awarded per build command

var ledger: RefCounted  # Ledger instance
var clock: RefCounted   # GameClock instance
var current_sovereign_id: String

var command_queue: Array[Dictionary] = []
var diplomacy_queue: Array[Dictionary] = []
var available_cp: int = 0

func _init(p_ledger: RefCounted, p_clock: RefCounted, p_sovereign_id: String) -> void:
	ledger = p_ledger
	clock = p_clock
	current_sovereign_id = p_sovereign_id

func start_turn() -> void:
	_run_cycle_a()
	_calculate_cp()

func _run_cycle_a() -> void:
	clock.advance_month()
	# Replace the clock dict wholesale to avoid field-level mutation of ledger state.
	ledger.game_clock = {"year": clock.year, "month": clock.month}

	var dominant: String = GameClockScript.element_for_month(clock.month)
	var season: String = GameClockScript.season_for_month(clock.month)

	ledger.log_event("SEASON_SHIFT", "The season is now %s, dominated by %s." % [season, dominant])

	var sovereign: Dictionary = ledger.get_officer(current_sovereign_id)
	if not sovereign.is_empty():
		var drift: float = EssenceScript.get_drift_multiplier(sovereign.essence, dominant)
		var eff_strategy: int = EssenceScript.get_effective_stat(sovereign.strategy, drift)
		ledger.log_event("ESSENCE_DRIFT", "%s feels the shift. Effective Strategy: %d (x%.2f)" % [current_sovereign_id, eff_strategy, drift])

func _calculate_cp() -> void:
	var sovereign: Dictionary = ledger.get_officer(current_sovereign_id)
	if sovereign.is_empty():
		available_cp = 5
		return

	var dominant: String = GameClockScript.element_for_month(clock.month)
	var drift: float = EssenceScript.get_drift_multiplier(sovereign.essence, dominant)
	var eff_strategy: int = EssenceScript.get_effective_stat(sovereign.strategy, drift)

	available_cp = max(5, eff_strategy / 10)
	ledger.log_event("RESOURCES", "Command Points generated: %d" % available_cp)

func queue_command(command_type: String, params: Dictionary, cost: int) -> bool:
	if available_cp >= cost:
		command_queue.append({"type": command_type, "params": params, "cost": cost})
		available_cp -= cost
		return true
	return false

func settle_turn() -> void:
	ledger.log_event("SETTLEMENT", "Resolving turn cycle D...")
	for cmd: Dictionary in command_queue:
		_execute_command(cmd)
	command_queue.clear()
	diplomacy_queue.clear()
	_run_cycle_d_yield()
	ledger.log_event("TURN_COMPLETE", "Turn %d.%d settled successfully." % [clock.year, clock.month])

func _execute_command(cmd: Dictionary) -> void:
	var city_name: String = cmd.params.get("city", "")
	var city: Dictionary = ledger.get_city(city_name)
	if city.is_empty():
		ledger.log_event("COMMAND_EXEC", "Unknown city for %s" % cmd.type)
		return

	match cmd.type:
		"BUILD_AG":
			city["agriculture"] = mini(100, city.get("agriculture", 0) + BUILD_GAIN)
			ledger.log_event("COMMAND_EXEC", "%s: agriculture → %d" % [city_name, city["agriculture"]])
		"BUILD_COM":
			city["commerce"] = mini(100, city.get("commerce", 0) + BUILD_GAIN)
			ledger.log_event("COMMAND_EXEC", "%s: commerce → %d" % [city_name, city["commerce"]])
		"BUILD_TECH":
			city["technology"] = mini(100, city.get("technology", 0) + BUILD_GAIN)
			ledger.log_event("COMMAND_EXEC", "%s: technology → %d" % [city_name, city["technology"]])
		"BUILD_ORD":
			city["order"] = mini(100, city.get("order", 0) + BUILD_GAIN)
			ledger.log_event("COMMAND_EXEC", "%s: order → %d" % [city_name, city["order"]])
		_:
			ledger.log_event("COMMAND_EXEC", "Unknown command %s" % cmd.type)

func _run_cycle_d_yield() -> void:
	var total_grain := 0
	var total_gold := 0
	for city_name: String in ledger.cities:
		var city: Dictionary = ledger.cities[city_name]
		var tech: int = city.get("technology", 0)
		var order: int = city.get("order", 0)
		var grain := CityEconomyScript.calc_grain_yield(city.get("agriculture", 0), tech)
		var gold := CityEconomyScript.calc_gold_yield(city.get("commerce", 0), tech)
		var loss := CityEconomyScript.calc_corruption_loss(order)
		total_grain += maxi(0, grain - loss)
		total_gold += maxi(0, gold - loss)
	ledger.resources["grain"] += total_grain
	ledger.resources["gold"] += total_gold
	ledger.log_event("YIELD", "Grain +%d, Gold +%d (total: %d / %d)" % [
		total_grain, total_gold, ledger.resources["grain"], ledger.resources["gold"]
	])
