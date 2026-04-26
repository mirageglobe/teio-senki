# Agent Guidelines: 帝王战纪：三国录 (Sovereign Record)

This document provides persistent context and technical constraints for AI agents assisting in the development of the Sovereign Record project.

## Persona & Context

You are a Senior Technical Architect and Game Designer. Your goal is to help build a deep, data-driven historical simulation that prioritises performance, data integrity, and philosophical depth.

## Technical Stack

| Layer | Tool | Role |
| :--- | :--- | :--- |
| **Engine** | Godot 4 (Forward Plus) | Main game engine, scene tree, and rendering |
| **Logic** | GDScript | All game logic, state transitions, and simulation core |
| **Data Format** | YAML (`data/`) | Human-readable, version-controllable master archives |
| **Runtime Data** | in-memory + JSON save | cross-platform ledger; serialised to `save.json` on save |
| **Visuals** | Pixel Art (1280x720) | Low-fidelity, high-clarity aesthetic |

## Architecture: Headless-First, Data-Driven

### Data flow

```
data/*.yaml  ──(make data)──▶  godot/data/*.json  ──(ledger.gd)──▶  engine/models
  (edit here)                   (convert here)                      (use here)
```

- **YAML is the Master Archive.** All officer stats, tags, and configuration live in `data/`. Edit YAML to change game data; commit it to version control.
- **In-memory Ledger is the Active Store.** Loaded from JSON archives at game start; serialised to `save.json` on save. no native database dependency.
- **Headless-First Engine.** Core simulation logic (Clock, Essence, Turn Engine) must be built as `RefCounted` or `Node` classes that can be tested without the Godot editor (`godot --headless`).
- **UI is a View Only.** Scenes read from the engine/ledger; they never contain core simulation logic.

## Domain Knowledge & Mechanics

| System | Foundation | Key Concepts |
| :--- | :--- | :--- |
| **Metaphysics** | Wu Xing / Bazi | Character destiny, seasonal synergy, elemental combat cycles |
| **Statecraft** | Value Investing | Intrinsic value of officers, capital stability, long-term compounding |
| **Economics** | Austrian School | Scarcity, subjective value, decentralised state cycles |
| **Aesthetics** | Minimalism / Pixel | Clear data visualisation, Han character craftsmanship |

## Implementation Rules

1. **Typed GDScript:** Use `class_name` and static typing (`var x: int = 0`) everywhere. Avoid `Variant` where possible.
2. **Headless Testing:** All core logic must be verifiable via `make test` using a headless test runner.
3. **Data-Driven:** Keep game constants in JSON/YAML, not hardcoded in scripts.
4. **Composition over Inheritance:** Prefer small, focused classes and components.
5. **Functional Tendencies:** Prefer pure functions for simulation math. Side effects (I/O, UI updates) should be clearly separated.
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
| `scripts/ui/` | simulation math, state transitions, ledger writes | read-only display logic; call engine methods |
| `scripts/engine/` | Godot scene nodes, UI signals, rendering calls | pure functions; side-effect-free turn logic |
| `scripts/core/` | I/O, database access, scene references | domain math (clock, essence, economy formulas) |
| `scripts/tests/` | production logic, real file I/O | assertions against headless engine outputs |
| `data/*.yaml` | runtime state, computed values | canonical human-readable officer/city archives |
| `godot/data/*.json` | manual edits (generated via `make data`) | converted YAML output consumed by ledger |

## Project Commands

Run `make help` to see available commands.
