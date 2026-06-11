package ledger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/models"
)

func writeTempData(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	officers := map[string]any{
		"officers": []models.Officer{
			{Name: "Cao Cao", Essence: "Metal", Tags: []string{"lord"}, Strategy: 90, Valour: 75, Governance: 85},
			{Name: "Liu Bei", Essence: "Wood", Tags: []string{"lord"}, Strategy: 80, Valour: 78, Governance: 82},
			{Name: "Xu Zhu", Essence: "Earth", Tags: []string{"general"}, Strategy: 30, Valour: 95, Governance: 20},
		},
	}
	cities := map[string]any{
		"cities": []models.City{
			{Name: "Luoyang", Agriculture: 80, Commerce: 85, Technology: 80, Order: 35},
			{Name: "Chengdu", Agriculture: 92, Commerce: 80, Technology: 75, Order: 78},
		},
	}

	writeJSON(t, filepath.Join(dir, "officers.json"), officers)
	writeJSON(t, filepath.Join(dir, "cities.json"), cities)
	return dir
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestNew(t *testing.T) {
	l := ledger.New()
	if l.Year != 189 || l.Month != 1 {
		t.Errorf("want year=189 month=1, got %d.%d", l.Year, l.Month)
	}
	if l.Officers == nil || l.Cities == nil {
		t.Error("maps must be non-nil")
	}
}

func TestLoadData(t *testing.T) {
	l := ledger.New()
	dir := writeTempData(t)
	if err := l.LoadData(dir); err != nil {
		t.Fatalf("LoadData: %v", err)
	}
	if len(l.Officers) != 3 {
		t.Errorf("want 3 officers, got %d", len(l.Officers))
	}
	if len(l.Cities) != 2 {
		t.Errorf("want 2 cities, got %d", len(l.Cities))
	}
}

func TestLoadData_MissingDir(t *testing.T) {
	l := ledger.New()
	if err := l.LoadData("/nonexistent/path"); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestLords(t *testing.T) {
	l := ledger.New()
	dir := writeTempData(t)
	if err := l.LoadData(dir); err != nil {
		t.Fatalf("LoadData: %v", err)
	}

	lords := l.Lords()
	if len(lords) != 2 {
		t.Fatalf("want 2 lords, got %d", len(lords))
	}
	// sorted by strategy descending: Cao Cao (90) > Liu Bei (80)
	if lords[0].Name != "Cao Cao" {
		t.Errorf("want first lord Cao Cao, got %s", lords[0].Name)
	}
	if lords[1].Name != "Liu Bei" {
		t.Errorf("want second lord Liu Bei, got %s", lords[1].Name)
	}
}

func TestGetSetCity(t *testing.T) {
	l := ledger.New()
	dir := writeTempData(t)
	if err := l.LoadData(dir); err != nil {
		t.Fatalf("LoadData: %v", err)
	}

	city, ok := l.GetCity("Luoyang")
	if !ok {
		t.Fatal("Luoyang not found")
	}
	city.Agriculture = 99
	l.SetCity("Luoyang", city)

	updated, _ := l.GetCity("Luoyang")
	if updated.Agriculture != 99 {
		t.Errorf("want agriculture=99, got %d", updated.Agriculture)
	}
}

func TestGetCity_Missing(t *testing.T) {
	l := ledger.New()
	_, ok := l.GetCity("Atlantis")
	if ok {
		t.Error("expected not found for unknown city")
	}
}

func TestGetOfficer(t *testing.T) {
	l := ledger.New()
	dir := writeTempData(t)
	if err := l.LoadData(dir); err != nil {
		t.Fatalf("LoadData: %v", err)
	}

	o, ok := l.GetOfficer("Cao Cao")
	if !ok {
		t.Fatal("Cao Cao not found")
	}
	if o.Strategy != 90 {
		t.Errorf("want strategy=90, got %d", o.Strategy)
	}
}

func TestSortedCities(t *testing.T) {
	l := ledger.New()
	dir := writeTempData(t)
	if err := l.LoadData(dir); err != nil {
		t.Fatalf("LoadData: %v", err)
	}

	cities := l.SortedCities()
	if len(cities) != 2 {
		t.Fatalf("want 2 cities, got %d", len(cities))
	}
	if cities[0].Name >= cities[1].Name {
		t.Errorf("cities not sorted: %s >= %s", cities[0].Name, cities[1].Name)
	}
}

func TestLog(t *testing.T) {
	l := ledger.New()
	l.Log("TEST", "hello")
	if len(l.Logs) != 1 {
		t.Fatalf("want 1 log, got %d", len(l.Logs))
	}
	e := l.Logs[0]
	if e.Type != "TEST" || e.Description != "hello" {
		t.Errorf("unexpected log entry: %+v", e)
	}
	if e.Year != 189 || e.Month != 1 {
		t.Errorf("log date mismatch: %d.%d", e.Year, e.Month)
	}
}

func TestState(t *testing.T) {
	l := ledger.New()
	l.Resources.Grain = 50
	l.Resources.Gold = 30
	l.Log("X", "test")

	state := l.State(7)
	if state.AvailableCP != 7 {
		t.Errorf("want CP=7, got %d", state.AvailableCP)
	}
	if state.Resources.Grain != 50 || state.Resources.Gold != 30 {
		t.Errorf("unexpected resources: %+v", state.Resources)
	}
	if len(state.Logs) != 1 {
		t.Errorf("want 1 log in snapshot, got %d", len(state.Logs))
	}
	// snapshot must be a copy — mutations must not affect ledger
	state.Logs[0].Type = "MUTATED"
	if l.Logs[0].Type == "MUTATED" {
		t.Error("State() must return a copy of logs, not a slice alias")
	}
}
