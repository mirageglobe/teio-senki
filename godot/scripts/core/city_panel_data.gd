class_name CityPanelData
extends RefCounted

const CityEconomyScript = preload("res://scripts/core/city_economy.gd")

# Pure function: derives all display data for a city detail panel from raw city dict.
static func format(city: Dictionary) -> Dictionary:
	var name_str: String = city.get("name", "Unknown")
	var chinese_str: String = city.get("chinese", "")
	var ag: int = city.get("agriculture", 0)
	var com: int = city.get("commerce", 0)
	var tech: int = city.get("technology", 0)
	var order: int = city.get("order", 0)

	return {
		"title": "%s %s" % [name_str, chinese_str] if chinese_str != "" else name_str,
		"region": city.get("region", "—"),
		"terrain": city.get("terrain", "—"),
		"faction": city.get("faction", "—"),
		"population": city.get("population", 0),
		"defense": city.get("defense", 0),
		"pillars": {
			"agriculture": ag,
			"commerce": com,
			"technology": tech,
			"order": order,
		},
		"grain_yield": CityEconomyScript.calc_grain_yield(ag, tech),
		"gold_yield": CityEconomyScript.calc_gold_yield(com, tech),
		"corruption_loss": CityEconomyScript.calc_corruption_loss(order),
	}
