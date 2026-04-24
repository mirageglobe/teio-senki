class_name CityEconomy
extends RefCounted

# Technology bonus ceiling: Fire generates Wood/Metal in Wu Xing, capped at 50% uplift.
const TECH_BONUS_MAX := 0.5

# Pure function: grain output (Wood pillar) amplified by technology (Fire generates Wood).
# Returns integer yield; minimum 0.
static func calc_grain_yield(agriculture: int, technology: int) -> int:
	var bonus := (float(technology) / 100.0) * TECH_BONUS_MAX
	return maxi(0, int(float(agriculture) * (1.0 + bonus)))

# Pure function: gold output (Metal pillar) amplified by technology (Fire melts/refines Metal).
# Returns integer yield; minimum 0.
static func calc_gold_yield(commerce: int, technology: int) -> int:
	var bonus := (float(technology) / 100.0) * TECH_BONUS_MAX
	return maxi(0, int(float(commerce) * (1.0 + bonus)))

# Pure function: corruption loss per turn (Earth pillar — order is the moat).
# order=100 → 0 loss; order=0 → 100 loss. Loss is a flat point drain on yields.
static func calc_corruption_loss(order: int) -> int:
	return maxi(0, 100 - clampi(order, 0, 100))
