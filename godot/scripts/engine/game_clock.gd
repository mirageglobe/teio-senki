class_name GameClock
extends RefCounted

# Bazi (Heavenly Stems and Earthly Branches)
const HEAVENLY_STEMS = ["Jia", "Yi", "Bing", "Ding", "Wu", "Ji", "Geng", "Xin", "Ren", "Gui"]
const EARTHLY_BRANCHES = ["Zi", "Chou", "Yin", "Mao", "Chen", "Si", "Wu", "Wei", "Shen", "You", "Xu", "Hai"]

# Elements associated with Stems (for reference)
const STEM_ELEMENTS = {
	"Jia": "Wood", "Yi": "Wood",
	"Bing": "Fire", "Ding": "Fire",
	"Wu": "Earth", "Ji": "Earth",
	"Geng": "Metal", "Xin": "Metal",
	"Ren": "Water", "Gui": "Water"
}

# Elements associated with Branches (Seasonal)
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
var month: int = 1 # 1-12
var total_months: int = 0 # Months since epoch AD 184

func _init(p_year: int = 189, p_month: int = 1):
	year = p_year
	month = p_month
	total_months = (year - 184) * 12 + (month - 1)

func advance_month():
	total_months += 1
	month += 1
	if month > 12:
		month = 1
		year += 1

func get_stem() -> String:
	# Year determines the stem (approximate for game epoch)
	return HEAVENLY_STEMS[(year - 184) % 10]

func get_branch() -> String:
	# Month determines the branch
	return EARTHLY_BRANCHES[(month - 1) % 12]

func get_dominant_element() -> String:
	return BRANCH_DATA[get_branch()]["element"]

func get_current_season() -> String:
	return BRANCH_DATA[get_branch()]["season"]

func get_display_date() -> String:
	return "Year %d, Month %d (%s %s)" % [year, month, get_stem(), get_branch()]
