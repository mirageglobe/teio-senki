class_name Essence
extends RefCounted

const ELEMENTS = ["Wood", "Fire", "Earth", "Metal", "Water"]

# Wu Xing generation cycle: each element feeds the next (木生火, 火生土, ...)
const GENERATION = {
	"Wood": "Fire", "Fire": "Earth", "Earth": "Metal", "Metal": "Water", "Water": "Wood"
}

# Wu Xing control cycle: each element suppresses the element two steps ahead (木克土, ...)
const CONTROL = {
	"Wood": "Earth", "Earth": "Water", "Water": "Fire", "Fire": "Metal", "Metal": "Wood"
}

const MULTIPLIERS = {
	"PEAK": 1.25, "NOURISHED": 1.15, "FEEDING": 1.10, "RESISTANT": 0.90, "SUPPRESSED": 0.80
}

# Pure function: returns drift multiplier based on officer's elemental root vs season's dominant element.
# PEAK = same element; NOURISHED = dominant generates officer; FEEDING = officer generates dominant;
# SUPPRESSED = dominant controls officer; RESISTANT = officer controls dominant.
static func get_drift_multiplier(officer_essence: String, dominant_element: String) -> float:
	var o_ess := officer_essence.capitalize()
	var d_ele := dominant_element.capitalize()

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

# Pure function: applies drift multiplier to a base stat, clamped to valid range [1, 100].
static func get_effective_stat(base_stat: int, multiplier: float) -> int:
	return clampi(int(round(float(base_stat) * multiplier)), 1, 100)
