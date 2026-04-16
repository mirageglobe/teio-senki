# 帝王战纪：三国录 (Sovereign Record: Three Kingdoms Ledger)

> "Sovereignty through the Ledger, Strategy through the Elements."

**帝王战纪：三国录** is a headless grand strategy engine and historical simulation built in Python. It uses a Terminal-First architecture — a high-performance TUI that feels like a modern retro war record.

## Vision

To create a definitive record of sovereignty, starting with the **Three Kingdoms era (AD 184–280)**. The engine is designed as a multi-era platform, capable of scaling to the Sengoku period, the Roman Republic, and beyond under the master brand **帝王战纪 (Sovereign Record)**.

## Technical Stack

| Layer | Tool | Role |
| :--- | :--- | :--- |
| Language | Python 3.11+ | Core logic, state management, world simulation |
| TUI | Textual | Terminal-first interface — widgets, layout, live updates |
| Data Integrity | Pydantic | Schema validation and security for all domain models |
| Ledger | SQLite | Local-first, append-style historical record of world state |
| Environment | uv | Dependency and virtual environment management |

## Architecture

- **Terminal-First:** Textual is the primary interface. No web or desktop frontend.
- **Pydantic models** are the single source of truth for all domain entities. Never pass raw dicts across boundaries.
- **SQLite Ledger:** Treat as append-only. World state is reconstructed from history, not mutated in place.
- **Headless engine:** Core simulation is importable and testable without launching the TUI.

---

## Game Mechanics

### Overview

The player assumes the role of a sovereign (君主) during the Three Kingdoms period (AD 184–280). Each turn represents one month. The goal is to build a lasting state — not through brute conquest, but through principled statecraft, elemental alignment, and the long-term compounding of institutional strength.

The engine is headless and pure: all simulation logic lives in `engine.py` and `clock.py` as side-effect-free functions. The TUI (`app.py`) is a thin shell over the engine. The SQLite Ledger (`ledger.py`) is the only place where state is written.

---

### Data Architecture

Two canonical sources feed the game at startup:

| Layer | File | Role |
| :--- | :--- | :--- |
| Master Archive | `data/officers.yaml`, `data/cities.yaml` | Human-readable source of truth; edited by hand |
| Active Ledger | `ledger.db` (SQLite) | Append-style runtime record; derived from the Archive |

On first launch, `main.py` seeds the Ledger from the YAML archives. Subsequent runs read only from SQLite. The Archive is never touched at runtime.

---

### Turn Structure

Each turn is a three-phase batch job. The world resolves before the player acts, and all changes are committed atomically to the Ledger at the end. Nothing is written until Phase C completes successfully.

#### Phase A — Market & Metaphysic Shift

The world updates automatically. The player has no control over this phase.

- The **Bazi clock** (天干地支) advances by one month. A new Heavenly Stem (天干) and Earthly Branch (地支) take effect.
- The **dominant Wu Xing element** shifts based on the new season: Wood (Spring), Fire (Summer), Metal (Autumn), Water (Winter), Earth (transitions).
- **Market signals** are generated from the dominant element and applied as flat deltas to every city's four economic pillars before the player acts.
- Officers' effective stats are recalculated live using the new **essence drift** multipliers.

Seasonal deltas applied to all cities each turn:

| Season | Element | AG | COM | TECH | ORD |
| :--- | :--- | :---: | :---: | :---: | :---: |
| Spring | Wood | +2 | 0 | +1 | 0 |
| Summer | Fire | +1 | +2 | +2 | -1 |
| Transition | Earth | 0 | 0 | 0 | +2 |
| Autumn | Metal | +3 | +1 | 0 | +1 |
| Winter | Water | -2 | -1 | 0 | 0 |

#### Phase B — Resource Allocation

The player issues **Commands** within a finite **Command Point (CP)** budget. Commands are entered via text in the TUI (planned: numbered selection).

- Only officers carrying the **lord** tag are eligible to serve as the active ruler.
- CP = `active_lord.strategy (after essence drift) ÷ 10`, minimum 5.
- Commands target a specific city by ID and cost 1–2 CP each.
- Commands are queued — nothing executes yet. The engine validates the CP budget and rejects any command that would exceed it.

| Command | Input | CP Cost | Effect |
| :--- | :--- | :---: | :--- |
| build_agriculture | `ag <city_id>` | 1 | +2 AG in target city |
| build_commerce | `com <city_id>` | 1 | +2 COM in target city |
| build_technology | `tech <city_id>` | 2 | +2 TECH in target city |
| build_order | `ord <city_id>` | 1 | +2 ORD in target city |
| build_defense | `def <city_id>` | 2 | +2 DEF in target city |

All command strings are lowercase. Input is case-insensitive and stripped before parsing.

#### Phase C — Settlement

The engine resolves all deltas and commits a single atomic SQLite transaction:

1. Apply Phase A market signal deltas to all cities (clamped to 1–100 per pillar).
2. Apply Phase B commands to their target cities (clamped to 1–100 per pillar).
3. Advance the game clock in the `game_clock` table.
4. Append all Phase A events and Phase C command results to the `ledger_log`.

No partial state is ever written. If the transaction fails, the Ledger is unchanged.

---

### Cities & Economic Pillars

The world contains **29 cities** seeded from `data/cities.yaml`, each with a fixed map position (x, y grid 0–19, 0–14), terrain type, population, faction, and five economic pillars:

| Pillar | Code | Domain | Notes |
| :--- | :--- | :--- | :--- |
| **Agriculture** | AG | Food / population growth | Peaks in Autumn; depressed in Winter |
| **Commerce** | COM | Trade / revenue | Peaks in Summer; slows in Winter |
| **Technology** | TECH | Innovation / infrastructure | Boosted by Fire element and Summer |
| **Order** | ORD | Stability / loyalty | Boosted by Earth (transitions); degraded by Summer unrest |
| **Defense** | DEF | Military fortification | Changed only by player commands |

All pillars are clamped to 1–100. Phase A applies seasonal deltas automatically each turn. Phase B commands apply targeted deltas to a single city. Both stack additively before the Phase C commit.

Terrain types affect future combat and logistics (not yet active):

| Terrain | Tags Favoured |
| :--- | :--- |
| Plain | Cavalry |
| Mountain | Engineer, Pioneer |
| Forest | — |
| River | Naval |
| Coast | Naval, Merchant |
| Pass | Engineer, Strategist |

---

### Officers (武将 / 文臣)

Officers are the core asset of any state. Each officer carries a **civil** or **military** tag and a dynamic Loyalty score affected by salary, treatment, rival offers, and elemental resonance with the sovereign's Bazi. The archive currently holds **158 historical officers** across Wei, Shu, Wu, and independent factions.

Recruiting an officer is an investment decision: price the candidate correctly, assess hidden value, and consider long-term compounding over short-term utility.

#### Officer Stats

Each officer is defined by five stats and an elemental Essence derived from their Bazi birth chart:

| Pillar | Domain | Definition |
| :--- | :--- | :--- |
| **Strategy** | Military / Macro | The Army Throughput metric. Higher Strategy allows for larger unit formations, better logistics, and superior tactical positioning on the grid. |
| **Valour** | Martial / Micro | The Personal Execution metric. Used exclusively for 1-to-1 duels and leading Vanguard charges to break enemy morale. |
| **Governance** | Economics / Fiscal | The Asset Management metric. Governs city tax yields, trade efficiency, and the compounding of infrastructure growth over time. |
| **Integrity** | Security / Loyalty | The System Resilience metric. Acts as the encryption for an officer's loyalty. A high Integrity officer is nearly immune to enemy bribery or plots. |
| **Essence** | Bazi / Metaphysics | The Elemental Root. Determines how the officer's performance fluctuates based on seasonal drift (e.g., a Water Essence officer thrives in Winter). |

Stats range from 1–100. Intrinsic values are hidden from rivals; the visible surface rating may differ. Seasonal drift applies a multiplier to all stats based on the Essence/season relationship:

#### Officer Tags

Tags represent an officer's specialist background. Each tag confers a percentage bonus to a specific pillar in the matching situation. An officer can hold multiple tags; bonuses stack independently per situation.

| Tag | Pillar Bonus | Situation |
| :--- | :--- | :--- |
| Lord | +20% Governance | Any policy action as head of state; also the source of Command Points |
| Naval | +25% Strategy | River / sea combat and logistics |
| Cavalry | +20% Valour | Open-terrain mounted charge |
| Vanguard | +15% Valour | Leading a charge or 1-on-1 duel |
| Scholar | +20% Governance | Academy, research, or civil administration |
| Diplomat | +20% Integrity | Alliance negotiation or espionage defence |
| Engineer | +20% Governance | Siege operation or infrastructure build |
| Merchant | +25% Governance | Active trade route |
| Strategist | +15% Strategy | All tactical positioning and logistics |
| Loyalist | +25% Integrity | Enemy bribery or plot attempt |
| Pioneer | +15% Governance | Governing a newly annexed territory |

Tags are stored in the Ledger and are part of every officer's permanent record. They do not change over time but can be discovered or hidden during intelligence operations.

| Essence | Peak Season | Suppressed Season |
| :--- | :--- | :--- |
| Wood | Spring | Autumn |
| Fire | Summer | Winter |
| Earth | Seasonal transitions | — |
| Metal | Autumn | Spring |
| Water | Winter | Summer |

---

### Five Elements (五行 Wu Xing)

All interactions in the world are governed by the elemental cycle:

```
Wood -> Fire -> Earth -> Metal -> Water -> Wood  (generating cycle)
Wood -x Earth -x Water -x Fire -x Metal -x Wood (controlling cycle)
```

- **Combat:** An army's dominant element interacts with terrain and season. A Fire general in summer on dry plains is at peak; the same general in winter rain is diminished.
- **Statecraft:** Policies carry elemental weight. A Metal-heavy administration (strict law, high taxes) controls but suppresses Wood (growth, agriculture). Balance is rewarded.
- **Officers:** An officer whose birth element is in a generating relationship with the sovereign's element gains loyalty and performance bonuses naturally.

---

### Seasons & Time

Each turn is one month. A year has twelve turns. Months are grouped into four seasons (春夏秋冬), each lasting three months, which affect all simulation systems:

| Season | Months | Element | Effect |
| :--- | :--- | :--- | :--- |
| Spring (春) | 1–3 | Wood | Agriculture output up; morale regenerates; ideal for expansion |
| Summer (夏) | 4–6 | Fire | Military campaigns peak; heat attrition in cold-affinity armies |
| Autumn (秋) | 7–9 | Metal | Harvest; resource banking; ideal for consolidation and law |
| Winter (冬) | 10–12 | Water | Movement slowed; supply costs up; diplomatic action favoured |

Monthly granularity matters: a campaign launched in month 8 (late Autumn) will enter Winter resupply pressure by month 10. Planning the timing of actions within a season is as important as the actions themselves.

---

### The Ledger (录)

Every action taken by the player is written to the Ledger — an append-only SQLite log. The Ledger is the historical record of your reign. Victory conditions are assessed against the Ledger, not live state.

**Currently recorded each turn:**

- Phase A market signals: season shift, grain signal, gold signal, tech pulse, order drift
- Phase C command results: which command was applied to which city and the resulting delta

**Planned:**

- Officer appointments, dismissals, deaths
- Battles: combatants, terrain, season, elemental modifiers, outcome
- Economic events: harvests, famines, trade agreements, fiscal policy
- Diplomatic exchanges: alliances, betrayals, marriages, tribute

The Log tab in the TUI displays the 100 most recent ledger entries in reverse chronological order.

---

### Statecraft & Economics

Economic management follows Austrian School principles — there is no central planner that optimises automatically:

- **Scarcity is real:** Grain, iron, timber, and coin are finite and location-bound. Moving them costs time and attrition.
- **Subjective value:** Trade prices between states fluctuate based on local supply and demand, not a fixed index.
- **Fiscal policy:** Tax rates affect short-term revenue and long-term loyalty. Overtaxing erodes the productive base.
- **Capital allocation:** The player must decide where to invest — walls, farms, academies, armories — knowing that compounding returns arrive slowly but durably.

---

### Victory & Legacy

There is no single win screen. Victory is measured across three dimensions recorded in the Ledger:

| Dimension | Measure |
| :--- | :--- |
| **Sovereignty** | Territory held at the era's end, weighted by strategic value |
| **Institutional Strength** | Officer loyalty, administrative depth, legal code quality |
| **Historical Judgment** | How later historians (the Ledger) record your decisions — moral, strategic, elemental |

A player who conquers through betrayal and scorched earth may hold the most land yet receive the weakest legacy score. The game rewards the sovereign who builds something that endures.

---

## Quick Start

```sh
make          # show available commands
make install  # install dependencies
make run      # launch the game
```

---

## Roadmap

### Planned

- **NPC event actors** — Introduce non-player characters such as physicians (e.g. Hua Tuo), Taoist sages, fortune tellers, and travelling merchants who appear as random events during Phase A. Each NPC triggers a contextual choice that affects officer stats, city conditions, or the elemental clock. NPCs are drawn from the officer archive and flagged by tag.
- **China map — unicode grid** — Render the campaign map in the TUI as a unicode character grid (UTF-8 box-drawing + Han radicals). Cities occupy fixed grid coordinates (x, y); terrain modifies the symbols. The map is navigable and serves as the primary strategic view.
- **Lowercase commands** — All in-game command strings use lowercase only (already enforced). Phase B input is case-insensitive and stripped before parsing.
- **Numbered selection UI** — Replace text-token command entry with a numbered selection system. The player presses 1–N to pick a command type, then 1–N to pick a target city, with CP cost shown inline. Reduces typing; compatible with keyboard-only TUI navigation.

### Under Consideration

- **Diplomacy phase** — Alliance proposals, tribute, hostage exchange, and espionage actions resolved between Phase A and Phase B.
- **Combat engine** — Grid-based army movement and battle resolution using strategy, valour, terrain bonuses, and elemental modifiers.
- **Officer recruitment** — Discover and recruit historical officers through events, diplomacy, or defection from rival factions.
- **Multi-era support** — Sengoku Japan, Roman Republic, and other historical periods under the master brand.

---

*Part of the **Sovereign Record** suite of historical simulations.*
