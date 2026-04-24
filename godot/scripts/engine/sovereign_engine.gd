class_name SovereignEngine
extends RefCounted

var ledger
var clock
var current_sovereign_id: String = "Cao Cao"

var command_queue: Array = []
var diplomacy_queue: Array = []
var available_cp: int = 0

func _init(p_ledger, p_clock):
	ledger = p_ledger
	clock = p_clock

func start_turn():
	_run_cycle_a()
	_calculate_cp()

func _run_cycle_a():
	clock.advance_month()
	ledger.game_clock.year = clock.year
	ledger.game_clock.month = clock.month
	
	var dominant = clock.get_dominant_element()
	var season = clock.get_current_season()
	
	ledger.log_event("SEASON_SHIFT", "The season is now %s, dominated by %s." % [season, dominant])
	
	var sovereign = ledger.get_officer(current_sovereign_id)
	if not sovereign.is_empty():
		# Using global load if Essence class name is not recognized
		var essence_cls = load("res://scripts/engine/essence.gd")
		var drift = essence_cls.get_drift_multiplier(sovereign.essence, dominant)
		var eff_strategy = essence_cls.get_effective_stat(sovereign.strategy, drift)
		ledger.log_event("ESSENCE_DRIFT", "%s feels the shift. Effective Strategy: %d (x%.2f)" % [current_sovereign_id, eff_strategy, drift])

func _calculate_cp():
	var sovereign = ledger.get_officer(current_sovereign_id)
	if sovereign.is_empty():
		available_cp = 5
		return
		
	var dominant = clock.get_dominant_element()
	var essence_cls = load("res://scripts/engine/essence.gd")
	var drift = essence_cls.get_drift_multiplier(sovereign.essence, dominant)
	var eff_strategy = essence_cls.get_effective_stat(sovereign.strategy, drift)
	
	available_cp = max(5, eff_strategy / 10)
	ledger.log_event("RESOURCES", "Command Points generated: %d" % available_cp)

func queue_command(command_type: String, params: Dictionary, cost: int):
	if available_cp >= cost:
		command_queue.append({"type": command_type, "params": params, "cost": cost})
		available_cp -= cost
		return true
	return false

func settle_turn():
	ledger.log_event("SETTLEMENT", "Resolving turn cycle D...")
	for cmd in command_queue:
		_execute_command(cmd)
	command_queue.clear()
	diplomacy_queue.clear()
	ledger.log_event("TURN_COMPLETE", "Turn %d.%d settled successfully." % [clock.year, clock.month])

func _execute_command(cmd: Dictionary):
	ledger.log_event("COMMAND_EXEC", "Executed %s with cost %d" % [cmd.type, cmd.cost], cmd.params)
