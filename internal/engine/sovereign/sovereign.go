// Package sovereign owns the 3-cycle turn engine: A (world update),
// B (command queuing), C (settlement). It does NOT own UI or I/O.
package sovereign

import (
	"fmt"

	"github.com/mirageglobe/teio-senki/internal/core/clock"
	"github.com/mirageglobe/teio-senki/internal/core/economy"
	"github.com/mirageglobe/teio-senki/internal/core/essence"
	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/models"
)

const buildGain = 5

// Engine drives the turn loop for a single sovereign.
type Engine struct {
	ledger       *ledger.Ledger
	clock        clock.State
	sovereignID  string
	commandQueue []models.Command
	availableCP  int
}

// New creates an Engine for the given sovereign.
func New(l *ledger.Ledger, sovereignID string) *Engine {
	return &Engine{
		ledger:      l,
		clock:       clock.NewAt(l.Year, l.Month),
		sovereignID: sovereignID,
	}
}

// StartTurn runs cycle A (world update) and calculates CP for cycle B.
func (e *Engine) StartTurn() {
	e.runCycleA()
	e.calculateCP()
}

// QueueCommand adds a command to the queue if CP allows.
// Returns true if accepted.
func (e *Engine) QueueCommand(cmd models.Command) bool {
	if e.availableCP < cmd.Cost {
		return false
	}
	e.commandQueue = append(e.commandQueue, cmd)
	e.availableCP -= cmd.Cost
	return true
}

// SettleTurn runs cycle C: executes commands then settles the economy.
// Returns the log entries produced this turn.
func (e *Engine) SettleTurn() []models.LogEntry {
	before := len(e.ledger.Logs)
	e.ledger.Log("SETTLEMENT", "resolving turn cycle C...")
	for _, cmd := range e.commandQueue {
		e.executeCommand(cmd)
	}
	e.commandQueue = e.commandQueue[:0]
	e.runCycleCSettle()
	e.ledger.Log("TURN_COMPLETE", fmt.Sprintf("turn %d.%d settled.", e.clock.Year, e.clock.Month))

	entries := e.ledger.Logs[before:]
	result := make([]models.LogEntry, len(entries))
	copy(result, entries)
	return result
}

// GetState returns a read-only snapshot for the UI.
func (e *Engine) GetState() models.GameState {
	return e.ledger.State(e.availableCP)
}

func (e *Engine) runCycleA() {
	e.clock = clock.Advance(e.clock)
	e.ledger.Year = e.clock.Year
	e.ledger.Month = e.clock.Month

	dominant := clock.ElementForMonth(e.clock.Month)
	season := clock.SeasonForMonth(e.clock.Month)
	e.ledger.Log("SEASON_SHIFT", fmt.Sprintf("season is %s, dominated by %s.", season, dominant))

	if officer, ok := e.ledger.GetOfficer(e.sovereignID); ok {
		drift := essence.DriftMultiplier(officer.Essence, dominant)
		effStrategy := essence.EffectiveStat(officer.Strategy, drift)
		e.ledger.Log("ESSENCE_DRIFT",
			fmt.Sprintf("%s: effective strategy %d (x%.2f)", e.sovereignID, effStrategy, drift))
	}
}

func (e *Engine) calculateCP() {
	officer, ok := e.ledger.GetOfficer(e.sovereignID)
	if !ok {
		e.availableCP = 5
		return
	}
	dominant := clock.ElementForMonth(e.clock.Month)
	drift := essence.DriftMultiplier(officer.Essence, dominant)
	effStrategy := essence.EffectiveStat(officer.Strategy, drift)
	e.availableCP = max(5, effStrategy/10)
	e.ledger.Log("RESOURCES", fmt.Sprintf("command points: %d", e.availableCP))
}

func (e *Engine) executeCommand(cmd models.Command) {
	cityName := cmd.Params["city"]
	city, ok := e.ledger.GetCity(cityName)
	if !ok {
		e.ledger.Log("COMMAND_EXEC", fmt.Sprintf("unknown city for %s", cmd.Type))
		return
	}
	switch cmd.Type {
	case "BUILD_AG":
		city.Agriculture = min(100, city.Agriculture+buildGain)
		e.ledger.SetCity(cityName, city)
		e.ledger.Log("COMMAND_EXEC", fmt.Sprintf("%s: agriculture → %d", cityName, city.Agriculture))
	case "BUILD_COM":
		city.Commerce = min(100, city.Commerce+buildGain)
		e.ledger.SetCity(cityName, city)
		e.ledger.Log("COMMAND_EXEC", fmt.Sprintf("%s: commerce → %d", cityName, city.Commerce))
	default:
		e.ledger.Log("COMMAND_EXEC", fmt.Sprintf("unknown command %s", cmd.Type))
	}
}

func (e *Engine) runCycleCSettle() {
	totalGrain, totalGold := 0, 0
	for _, city := range e.ledger.Cities {
		grain := economy.GrainYield(city.Agriculture, city.Technology)
		gold := economy.GoldYield(city.Commerce, city.Technology)
		loss := economy.CorruptionLoss(city.Order)
		if net := grain - loss; net > 0 {
			totalGrain += net
		}
		if net := gold - loss; net > 0 {
			totalGold += net
		}
	}
	e.ledger.Resources.Grain += totalGrain
	e.ledger.Resources.Gold += totalGold
	e.ledger.Log("YIELD", fmt.Sprintf("grain +%d, gold +%d (total: %d / %d)",
		totalGrain, totalGold, e.ledger.Resources.Grain, e.ledger.Resources.Gold))
}
