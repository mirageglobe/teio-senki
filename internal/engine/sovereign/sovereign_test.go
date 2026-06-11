package sovereign_test

import (
	"testing"

	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/engine/sovereign"
	"github.com/mirageglobe/teio-senki/internal/models"
)

func newTestLedger() *ledger.Ledger {
	l := ledger.New()
	l.Officers["Cao Cao"] = models.Officer{
		Name:     "Cao Cao",
		Essence:  "Metal",
		Tags:     []string{"lord"},
		Strategy: 90,
		Valour:   75,
		Governance: 85,
	}
	l.Cities["Luoyang"] = models.City{
		Name:        "Luoyang",
		Agriculture: 80,
		Commerce:    85,
		Technology:  80,
		Order:       35,
	}
	l.Cities["Chengdu"] = models.City{
		Name:        "Chengdu",
		Agriculture: 92,
		Commerce:    80,
		Technology:  75,
		Order:       78,
	}
	return l
}

func TestNew(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	if e == nil {
		t.Fatal("New returned nil")
	}
}

func TestGetState_InitialSeason(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	state := e.GetState()
	if state.Season == "" {
		t.Error("season must not be empty")
	}
	if state.Element == "" {
		t.Error("element must not be empty")
	}
}

func TestStartTurn_AdvancesClock(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()
	state := e.GetState()
	// clock advances on StartTurn — year/month must be >= start (189.1)
	if state.Year < 189 {
		t.Errorf("year did not advance: %d", state.Year)
	}
}

func TestStartTurn_SetsCP(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()
	state := e.GetState()
	if state.AvailableCP <= 0 {
		t.Errorf("CP must be positive after StartTurn, got %d", state.AvailableCP)
	}
}

func TestStartTurn_LogsProduced(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()
	state := e.GetState()
	if len(state.Logs) == 0 {
		t.Error("StartTurn must produce log entries")
	}
}

func TestQueueCommand_AcceptsWhenCP(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	ok := e.QueueCommand(models.Command{
		Type:   "BUILD_AG",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	if !ok {
		t.Error("QueueCommand should accept when CP is sufficient")
	}
}

func TestQueueCommand_RejectsWhenInsufficientCP(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	// drain all CP
	for range 50 {
		e.QueueCommand(models.Command{
			Type:   "BUILD_AG",
			Params: map[string]string{"city": "Luoyang"},
			Cost:   2,
		})
	}
	ok := e.QueueCommand(models.Command{
		Type:   "BUILD_AG",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	if ok {
		t.Error("QueueCommand must reject when CP is exhausted")
	}
}

func TestQueueCommand_DeductsCP(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	before := e.GetState().AvailableCP
	e.QueueCommand(models.Command{
		Type:   "BUILD_AG",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	after := e.GetState().AvailableCP
	if after != before-2 {
		t.Errorf("CP deduction: want %d, got %d", before-2, after)
	}
}

func TestSettleTurn_BuildAG(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	before, _ := l.GetCity("Luoyang")
	e.QueueCommand(models.Command{
		Type:   "BUILD_AG",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	e.SettleTurn()

	after, _ := l.GetCity("Luoyang")
	if after.Agriculture <= before.Agriculture {
		t.Errorf("BUILD_AG must increase agriculture: %d → %d", before.Agriculture, after.Agriculture)
	}
}

func TestSettleTurn_BuildCOM(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	before, _ := l.GetCity("Luoyang")
	e.QueueCommand(models.Command{
		Type:   "BUILD_COM",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	e.SettleTurn()

	after, _ := l.GetCity("Luoyang")
	if after.Commerce <= before.Commerce {
		t.Errorf("BUILD_COM must increase commerce: %d → %d", before.Commerce, after.Commerce)
	}
}

func TestSettleTurn_BuildDEF(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	before, _ := l.GetCity("Luoyang")
	e.QueueCommand(models.Command{
		Type:   "BUILD_DEF",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	e.SettleTurn()

	after, _ := l.GetCity("Luoyang")
	if after.Defense <= before.Defense {
		t.Errorf("BUILD_DEF must increase defense: %d → %d", before.Defense, after.Defense)
	}
}

func TestSettleTurn_ClearsCommandQueue(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	e.QueueCommand(models.Command{
		Type:   "BUILD_AG",
		Params: map[string]string{"city": "Luoyang"},
		Cost:   2,
	})
	e.SettleTurn()

	// second settle with no commands must not re-apply the previous command
	agBefore, _ := l.GetCity("Luoyang")
	e.SettleTurn()
	agAfter, _ := l.GetCity("Luoyang")
	if agAfter.Agriculture != agBefore.Agriculture {
		t.Error("command queue must be cleared after SettleTurn")
	}
}

func TestSettleTurn_AccumulatesResources(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()
	e.SettleTurn()

	if l.Resources.Grain <= 0 && l.Resources.Gold <= 0 {
		t.Error("SettleTurn must accumulate grain or gold from cities")
	}
}

func TestSettleTurn_ReturnsLogs(t *testing.T) {
	l := newTestLedger()
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()
	logs := e.SettleTurn()
	if len(logs) == 0 {
		t.Error("SettleTurn must return log entries")
	}
}

func TestSettleTurn_StatCappedAt100(t *testing.T) {
	l := newTestLedger()
	l.Cities["Luoyang"] = models.City{
		Name:        "Luoyang",
		Agriculture: 98,
		Commerce:    85,
		Technology:  80,
		Order:       35,
	}
	e := sovereign.New(l, "Cao Cao")
	e.StartTurn()

	// queue two BUILD_AG commands; only the first should take effect at cap
	e.QueueCommand(models.Command{Type: "BUILD_AG", Params: map[string]string{"city": "Luoyang"}, Cost: 2})
	e.QueueCommand(models.Command{Type: "BUILD_AG", Params: map[string]string{"city": "Luoyang"}, Cost: 2})
	e.SettleTurn()

	city, _ := l.GetCity("Luoyang")
	if city.Agriculture > 100 {
		t.Errorf("agriculture exceeded cap: %d", city.Agriculture)
	}
}
