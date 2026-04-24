class_name SovereignEngine
extends RefCounted

# Explicit preloads so class_name globals are not required — safe for headless -s mode.
const EssenceScript = preload("res://scripts/core/essence.gd")
const GameClockScript = preload("res://scripts/core/game_clock.gd")

var ledger: RefCounted  # Ledger instance
var clock: RefCounted   # GameClock instance
var current_sovereign_id: String = "Cao Cao"

var command_queue: Array = []
var diplomacy_queue: Array = []
var available_cp: int = 0

func _init(p_ledger: RefCounted, p_clock: RefCounted) -> void:
	ledger = p_ledger
	clock = p_clock

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
	ledger.log_event("TURN_COMPLETE", "Turn %d.%d settled successfully." % [clock.year, clock.month])

func _execute_command(cmd: Dictionary) -> void:
	ledger.log_event("COMMAND_EXEC", "Executed %s with cost %d" % [cmd.type, cmd.cost], cmd.params)
