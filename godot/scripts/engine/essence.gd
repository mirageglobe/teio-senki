class_name Essence
extends RefCounted

const ELEMENTS = ["Wood", "Fire", "Earth", "Metal", "Water"]

const GENERATION = {
	"Wood": "Fire", "Fire": "Earth", "Earth": "Metal", "Metal": "Water", "Water": "Wood"
}

const CONTROL = {
	"Wood": "Earth", "Earth": "Water", "Water": "Fire", "Fire": "Metal", "Metal": "Wood"
}

const MULTIPLIERS = {
	"PEAK": 1.25, "NOURISHED": 1.15, "FEEDING": 1.10, "RESISTANT": 0.90, "SUPPRESSED": 0.80
}

static func get_drift_multiplier(officer_essence: String, dominant_element: String) -> float:
	var o_ess = officer_essence.capitalize()
	var d_ele = dominant_element.capitalize()
	
	if o_ess == d_ele:
		return MULTIPLIERS["PEAK"]
	
	if GENERATION.get(d_ele) == o_ess:
		return MULTIPLIERS["NOURISHED"]
	
	if GENERATION.get(o_ess) == d_ele:
		return MULTIPLIERS["FEEDING"]
	
	if CONTROL.get(d_ele) == o_ess:
		return MULTIPLIERS["SUPPRESSED"]
	
	if CONTROL.get(o_ess) == d_ele:
		return MULTIPLIERS["RESISTANT"]
	
	return 1.0

static func get_effective_stat(base_stat: int, multiplier: float) -> int:
	return clampi(int(round(float(base_stat) * multiplier)), 1, 100)
