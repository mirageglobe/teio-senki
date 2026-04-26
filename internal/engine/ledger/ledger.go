// Package ledger owns the in-memory game state. It is the single source of
// truth during a play session. It does NOT own simulation logic or UI.
package ledger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mirageglobe/teio-senki/internal/models"
)

// Ledger holds all runtime game state.
type Ledger struct {
	Officers  map[string]models.Officer
	Cities    map[string]models.City
	Year      int
	Month     int
	Resources models.Resources
	Logs      []models.LogEntry
}

// New returns a Ledger initialised to the default start state (AD 189.1).
func New() *Ledger {
	return &Ledger{
		Officers: map[string]models.Officer{},
		Cities:   map[string]models.City{},
		Year:     189,
		Month:    1,
	}
}

// LoadData reads officers.json and cities.json from the given directory.
func (l *Ledger) LoadData(dir string) error {
	if err := l.loadOfficers(filepath.Join(dir, "officers.json")); err != nil {
		return err
	}
	return l.loadCities(filepath.Join(dir, "cities.json"))
}

func (l *Ledger) loadOfficers(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("officers: %w", err)
	}
	var wrapper struct {
		Officers []models.Officer `json:"officers"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("officers parse: %w", err)
	}
	for _, o := range wrapper.Officers {
		l.Officers[o.Name] = o
	}
	return nil
}

func (l *Ledger) loadCities(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cities: %w", err)
	}
	var wrapper struct {
		Cities []models.City `json:"cities"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("cities parse: %w", err)
	}
	for _, c := range wrapper.Cities {
		l.Cities[c.Name] = c
	}
	return nil
}

// Lords returns all officers with the "lord" tag, sorted by strategy descending.
func (l *Ledger) Lords() []models.Officer {
	var result []models.Officer
	for _, o := range l.Officers {
		for _, tag := range o.Tags {
			if tag == "lord" {
				result = append(result, o)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Strategy > result[j].Strategy
	})
	return result
}

// Log appends an event entry.
func (l *Ledger) Log(entryType, description string) {
	l.Logs = append(l.Logs, models.LogEntry{
		Year:        l.Year,
		Month:       l.Month,
		Type:        entryType,
		Description: description,
	})
}

// GetOfficer returns an officer by name and whether it was found.
func (l *Ledger) GetOfficer(name string) (models.Officer, bool) {
	o, ok := l.Officers[name]
	return o, ok
}

// GetCity returns a city by name and whether it was found.
func (l *Ledger) GetCity(name string) (models.City, bool) {
	c, ok := l.Cities[name]
	return c, ok
}

// SetCity writes a city back into the ledger.
func (l *Ledger) SetCity(name string, c models.City) {
	l.Cities[name] = c
}

// State returns a read-only snapshot for the UI and Lua layers.
func (l *Ledger) State(availableCP int) models.GameState {
	logs := make([]models.LogEntry, len(l.Logs))
	copy(logs, l.Logs)
	return models.GameState{
		Year:        l.Year,
		Month:       l.Month,
		Resources:   l.Resources,
		AvailableCP: availableCP,
		Logs:        logs,
	}
}
