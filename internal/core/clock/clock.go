// Package clock owns the Bazi-based calendar: heavenly stems, earthly branches,
// dominant element and season per month. It does NOT own ledger state or I/O.
package clock

import "fmt"

var stems = [10]string{"Jia", "Yi", "Bing", "Ding", "Wu", "Ji", "Geng", "Xin", "Ren", "Gui"}
var branches = [12]string{"Zi", "Chou", "Yin", "Mao", "Chen", "Si", "Wu", "Wei", "Shen", "You", "Xu", "Hai"}

type branchInfo struct {
	element string
	season  string
}

var branchTable = map[string]branchInfo{
	"Zi":   {"Water", "Winter"},
	"Chou": {"Earth", "Transition"},
	"Yin":  {"Wood", "Spring"},
	"Mao":  {"Wood", "Spring"},
	"Chen": {"Earth", "Transition"},
	"Si":   {"Fire", "Summer"},
	"Wu":   {"Fire", "Summer"},
	"Wei":  {"Earth", "Transition"},
	"Shen": {"Metal", "Autumn"},
	"You":  {"Metal", "Autumn"},
	"Xu":   {"Earth", "Transition"},
	"Hai":  {"Water", "Winter"},
}

// State holds the current calendar position.
type State struct {
	Year        int
	Month       int // 1–12
	TotalMonths int // months elapsed since epoch AD 184
}

// New returns a State initialised to the default start date (AD 189, Month 1).
func New() State {
	return NewAt(189, 1)
}

// NewAt constructs a State for an arbitrary year/month.
func NewAt(year, month int) State {
	return State{
		Year:        year,
		Month:       month,
		TotalMonths: (year-184)*12 + (month - 1),
	}
}

// Advance returns a new State advanced by one month. Pure function.
func Advance(s State) State {
	next := s
	next.TotalMonths++
	next.Month++
	if next.Month > 12 {
		next.Month = 1
		next.Year++
	}
	return next
}

// StemForYear returns the heavenly stem for the given year.
func StemForYear(year int) string {
	return stems[(year-184)%10]
}

// BranchForMonth returns the earthly branch for a given month (1-indexed).
func BranchForMonth(month int) string {
	return branches[(month-1)%12]
}

// ElementForMonth returns the dominant Wu Xing element for the month.
func ElementForMonth(month int) string {
	return branchTable[BranchForMonth(month)].element
}

// SeasonForMonth returns the agricultural season for the month.
func SeasonForMonth(month int) string {
	return branchTable[BranchForMonth(month)].season
}

// DisplayDate formats a date as a human-readable string.
func DisplayDate(year, month int) string {
	return fmt.Sprintf("Year %d, Month %d (%s %s)", year, month, StemForYear(year), BranchForMonth(month))
}
