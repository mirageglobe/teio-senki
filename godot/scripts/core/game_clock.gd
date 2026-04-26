class_name GameClock
extends RefCounted

const HEAVENLY_STEMS = ["Jia", "Yi", "Bing", "Ding", "Wu", "Ji", "Geng", "Xin", "Ren", "Gui"]
const EARTHLY_BRANCHES = ["Zi", "Chou", "Yin", "Mao", "Chen", "Si", "Wu", "Wei", "Shen", "You", "Xu", "Hai"]

const STEM_ELEMENTS = {
	"Jia": "Wood", "Yi": "Wood",
	"Bing": "Fire", "Ding": "Fire",
	"Wu": "Earth", "Ji": "Earth",
	"Geng": "Metal", "Xin": "Metal",
	"Ren": "Water", "Gui": "Water"
}

# Month → earthly branch data; branch determines dominant element and agricultural season.
const BRANCH_DATA = {
	"Zi": {"element": "Water", "season": "Winter"},
	"Chou": {"element": "Earth", "season": "Transition"},
	"Yin": {"element": "Wood", "season": "Spring"},
	"Mao": {"element": "Wood", "season": "Spring"},
	"Chen": {"element": "Earth", "season": "Transition"},
	"Si": {"element": "Fire", "season": "Summer"},
	"Wu": {"element": "Fire", "season": "Summer"},
	"Wei": {"element": "Earth", "season": "Transition"},
	"Shen": {"element": "Metal", "season": "Autumn"},
	"You": {"element": "Metal", "season": "Autumn"},
	"Xu": {"element": "Earth", "season": "Transition"},
	"Hai": {"element": "Water", "season": "Winter"}
}

var year: int = 189
var month: int = 1
var total_months: int = 0  # months elapsed since epoch AD 184

func _init(p_year: int = 189, p_month: int = 1) -> void:
	year = p_year
	month = p_month
	total_months = (year - 184) * 12 + (month - 1)

# --- pure static functions (side-effect-free) ---

# Returns a new state dictionary with time advanced by one month.
# Prefer these in cycle A so turn resolution remains a pure transformation.
static func advance_state(state: Dictionary) -> Dictionary:
	var next := state.duplicate()
	next.total_months += 1
	next.month += 1
	if next.month > 12:
		next.month = 1
		next.year += 1
	return next

static func stem_for_year(p_year: int) -> String:
	return HEAVENLY_STEMS[(p_year - 184) % 10]

static func branch_for_month(p_month: int) -> String:
	return EARTHLY_BRANCHES[(p_month - 1) % 12]

static func element_for_month(p_month: int) -> String:
	return BRANCH_DATA[branch_for_month(p_month)]["element"]

static func season_for_month(p_month: int) -> String:
	return BRANCH_DATA[branch_for_month(p_month)]["season"]

static func display_date(p_year: int, p_month: int) -> String:
	return "Year %d, Month %d (%s %s)" % [p_year, p_month, stem_for_year(p_year), branch_for_month(p_month)]

# --- instance methods (delegate to pure statics) ---

func advance_month() -> void:
	# Self-reference via class name breaks under dynamic load; call statics directly.
	var next := advance_state({"year": year, "month": month, "total_months": total_months})
	year = next.year
	month = next.month
	total_months = next.total_months

func get_stem() -> String:
	return stem_for_year(year)

func get_branch() -> String:
	return branch_for_month(month)

func get_dominant_element() -> String:
	return element_for_month(month)

func get_current_season() -> String:
	return season_for_month(month)

func get_display_date() -> String:
	return display_date(year, month)
