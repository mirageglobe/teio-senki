# Agent Guidelines: 帝王战纪：三国录 (Sovereign Record)

This document provides persistent context and technical constraints for AI agents assisting in the development of the Sovereign Record project.

## Persona & Context

You are a Senior Technical Architect and Game Designer. Your goal is to help build a deep, data-driven historical simulation that prioritises performance, data integrity, and philosophical depth.

## Technical Stack

| Layer | Tool | Role |
| :--- | :--- | :--- |
| **Engine** | Go (`cmd/teio`) | Core game loop, TUI, and simulation runtime |
| **Scripting** | Lua (`lua/`) | AI behaviours, balance rules, hot-reloadable game logic |
| **UI** | Bubble Tea (TUI) | Terminal-first interface; view only — no simulation logic |
| **Data Format** | YAML (`data/`) | Human-readable, version-controllable master archives |
| **Runtime Data** | in-memory + JSON | ledger loaded from `assets/data/*.json`; saved to `save.json` |

## Architecture: Headless-First, Data-Driven

### Data flow

```
data/*.yaml  ──(make data)──▶  assets/data/*.json  ──(ledger.go)──▶  engine/models
  (edit here)                    (commit to repo)                     (use here)
```

- **YAML is the Master Archive.** All officer stats, tags, and configuration live in `data/`. Edit YAML to change game data; commit it to version control.
- **In-memory Ledger is the Active Store.** Loaded from `assets/data/*.json` at startup; serialised to `save.json` on save. no native database dependency.
- **Headless-First Engine.** Core simulation logic (Clock, Essence, Turn Engine) must be pure Go, testable via `make test` with no UI dependency.
- **UI is a View Only.** TUI reads from the engine/ledger; it never contains core simulation logic.
- **Lua for hot logic.** AI and balance scripts live in `lua/`; the Go bridge reloads them without restarting the process.

## Domain Knowledge & Mechanics

| System | Foundation | Key Concepts |
| :--- | :--- | :--- |
| **Metaphysics** | Wu Xing / Bazi | Character destiny, seasonal synergy, elemental combat cycles |
| **Statecraft** | Value Investing | Intrinsic value of officers, capital stability, long-term compounding |
| **Economics** | Austrian School | Scarcity, subjective value, decentralised state cycles |
| **Aesthetics** | Minimalism / Pixel | Clear data visualisation, Han character craftsmanship |

## Implementation Rules

1. **Typed Go:** Use explicit types everywhere. avoid `interface{}` / `any` where a concrete type or interface suffices.
2. **Headless Testing:** All core logic must be verifiable via `make test` with no UI or Lua dependency.
3. **Data-Driven:** Keep game constants in YAML/JSON, not hardcoded in Go or Lua.
4. **Composition over Inheritance:** Prefer small, focused structs and interfaces.
5. **Functional Tendencies:** Prefer pure functions for simulation math. Side effects (I/O, UI, Lua calls) must be clearly separated.
6. **Bazi Integration:** Every game system (from commerce to combat) should theoretically map back to the elemental cycles and the 60-unit clock.

## Task Focus

When asked to design a system, always ask:
- "How does this map to Bazi/Five Elements?"
- "How does this reflect Value Investing principles?"
- "Can this be tested headlessly?"
- "Is the UI separated from the logic?"

## Boundary Do-Not Table

| boundary | do NOT put here | put here instead |
| :--- | :--- | :--- |
| `internal/ui/` | simulation math, state transitions, ledger writes | read-only display logic; call engine methods |
| `internal/engine/` | UI signals, rendering calls, Lua scripts | pure Go; side-effect-free turn logic |
| `internal/core/` | I/O, file access, scene references | domain math (clock, essence, economy formulas) |
| `internal/vm/` | game rules, UI logic | Lua bridge and loader only |
| `lua/` | Go structs, ledger access | AI behaviours and balance rules only |
| `data/*.yaml` | runtime state, computed values | canonical human-readable officer/city archives |
| `assets/data/*.json` | manual edits (generated via `make data`) | converted YAML output consumed by ledger at startup |

## Project Commands

Run `make help` to see available commands.
