// Package models defines shared data types used across all packages.
// It does NOT own business logic — that belongs in core and engine packages.
package models

// Officer is a historical figure with elemental and stat attributes.
type Officer struct {
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Essence    string   `json:"essence"`
	Tags       []string `json:"tags"`
	Strategy   int      `json:"strategy"`
	Valour     int      `json:"valour"`
	Governance int      `json:"governance"`
	Integrity  int      `json:"integrity"`
	Loyalty    int      `json:"loyalty"`
}

// City is a controllable settlement with economic pillars.
type City struct {
	Name        string `json:"name"`
	Chinese     string `json:"chinese,omitempty"`
	Region      string `json:"region,omitempty"`
	Terrain     string `json:"terrain,omitempty"`
	X           int    `json:"x,omitempty"`
	Y           int    `json:"y,omitempty"`
	Population  int    `json:"population,omitempty"`
	Defense     int    `json:"defense,omitempty"`
	Faction     string `json:"faction,omitempty"`
	Agriculture int    `json:"agriculture"`
	Commerce    int    `json:"commerce"`
	Technology  int    `json:"technology"`
	Order       int    `json:"order"`
}

// Resources holds accumulated strategic resources.
type Resources struct {
	Grain int
	Gold  int
}

// LogEntry is an append-only event record.
type LogEntry struct {
	Year        int
	Month       int
	Type        string
	Description string
}

// Command is a queued player or AI action.
type Command struct {
	Type   string
	Params map[string]string
	Cost   int
}

// GameState is a read-only snapshot consumed by UI and Lua layers.
// Neither the UI nor Lua scripts may write back to the engine via this struct.
type GameState struct {
	Year        int
	Month       int
	Resources   Resources
	AvailableCP int
	Logs        []LogEntry
}
