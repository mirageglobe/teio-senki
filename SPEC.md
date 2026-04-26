# spec: 帝王战纪：三国录

> "Sovereignty through the Ledger, Strategy through the Elements."

## tldr
A headless, turn-based grand strategy engine set in the Three Kingdoms era. Architecture prioritises **simulation purity (headless-first)** and **cross-platform portability (in-memory ledger + JSON persistence)** over visual spectacle.

---

## project structure

```text
godot/
├── data/               # Converted JSON game data
├── scenes/             # Godot UI/Visual scenes
├── scripts/
│   ├── engine/         # Headless core: Clock, Essence, Ledger, Engine
│   ├── tests/          # Headless test runner & automated suite
│   └── ui/             # Godot UI scripts (Views only)
└── project.godot
data/                   # YAML Master Archive (Source of Truth)
```

## development principles

### simplicity first
- **flat over relational** — game state lives in flat dictionaries on domain models. no row-lifecycle tables, no FK references, no separate join records.
- **sequential over transactional** — apply deltas in order. if something breaks, fix the bug. no snapshot/restore, no rollback machinery.
- **no new pillar systems** — AG/COM/TECH/ORD are active in yield formulas (set by scenario data). DEF is active for sieges. player BUILD commands for TECH and ORD are deferred until the core loop is balanced.
- **static before dynamic** — population is a static recruitment cap. dynamic growth/decline is a polish feature, not a foundation.
- **distance over pathfinding** — supply is a Chebyshev distance check, not graph traversal. simple rules first.
- **territory wins** — victory is one faction holding all cities. multi-dimension scoring is deferred.

### ease of development
- **headless-first** — all engine logic is testable via `make test` without the Godot editor. UI is view-only.
- **vertical slice** — AD 189 scenario only. one lord tag subset (~50 officers). do not balance other scenarios until the game loop is playable end-to-end.
- **no just-in-case** — do not implement a feature until the turn engine needs it. espionage waits until diplomacy works; population dynamics wait until the economy is stable.
- **auto-resolve first** — no tactical grid in V1. the auto-resolve math formula is the battle system.
- **test before polish** — verify engine logic with placeholder Godot UI before committing to pixel art assets.

## scope containment (no-drift)
- **Engine/UI Separation**: Logic MUST be testable in `make test` without initializing the UI.
- **Ledger Boundary**: If a feature doesn't write state to the ledger, it is not part of the core simulation.
- **No Tactical Grid (V1)**: The grid battle is deferred to expansions. Stick to the auto-resolve formula.
- **Single Scenario**: Development is locked to **AD 189 (Dong Zhuo)**. Do not balance other scenarios until the game is playable.
- **No "Just-in-case"**: Do not implement features until the turn engine needs them.

---

## overview

**帝王战纪：三国录** is a turn-based grand strategy simulation set during China's Three Kingdoms era (AD 184–280), beginning at the Yellow Turban Rebellion epoch. the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft, elemental alignment, and long-term institutional building.

victory is measured by territory — one faction controls all cities. institutional and legacy scoring are deferred to post-release.

---

## expectations

- **low pixel art aesthetic** — sprites, tiles, and UI elements use a constrained pixel palette. no high-resolution assets. art style favours clarity and period character over visual complexity.
- **speed first** — all screens must feel instant. turn resolution, scene transitions, and UI interactions target < 100ms response. no loading screens between strategic map and city/army overlays.
- **keyboard-driven** — all core actions reachable without a mouse. mouse/touch supported as secondary input.
- **minimal UI chrome** — information density over decoration. panels appear on demand; the map is always the primary surface.
- **no tutorials** — the game communicates through design, tooltips, and the ledger log. no guided onboarding flow.

---

## architecture

### principles

- **headless engine**: all simulation logic (engine, clock) is side-effect-free and testable without the frontend.
- **typed domain models**: single source of truth for all entities. raw/untyped data never crosses module boundaries.
- **mutable current state + chronicle log**: in-memory dictionaries store live game state. `ledger_log` is an append-only array for history, narrative, and victory scoring — not the primary state store. state is serialised to JSON on save.
- **YAML archives**: human-readable canonical data (`data/officers.yaml`, `data/cities.yaml`). read once at first-run seed; never touched at runtime.

### module responsibilities

| module | role |
| :--- | :--- |
| main | entry point: loads JSON archives into ledger on first run, launches frontend |
| models | typed domain models: Officer, City, Army, Element, Tag, Terrain, and bonus lookup tables |
| clock | bazi calendar: heavenly stem (天干) / earthly branch (地支) tracking, essence drift calculation |
| engine | three-cycle turn processor: command validation, seasonal deltas, CP budget |
| battle | tactical battle resolver: grid movement, combat, duels, siege, morale |
| diplomacy | inter-faction relations: alliances, espionage, tribute, bribery |
| events | event generator: random events, historical triggers, NPC appearances |
| ledger | in-memory state store: officers, cities, clock, resources, append-only log; serialises to JSON on save |
| archive | YAML loader: parses human-readable config into validated domain models |
| frontend | Godot 4 scenes: splash, strategic map, city, officer, army, battle, diplomacy, ledger, victory |

### data flow

```
data/*.yaml  ──(make data)──▶  godot/data/*.json  ──(ledger.load_data)──▶  in-memory ledger
                                                                                    │
    frontend ──> engine (cycle A/B/C) ──────────────────────> ledger ──────────────┘
             └──> battle / diplomacy / events                        │
                                                         (settle_turn; sequential delta apply)
                                                                     │
                                                              save ──▶  save.json
```

---

## domain models

### Officer

| field | type | notes |
| :--- | :--- | :--- |
| id | string | unique identifier |
| name | string | historical name |
| title | string | honorific / rank |
| essence | Element | elemental root; drives seasonal drift multiplier |
| strategy | int 1–100 | army throughput, CP generation |
| valour | int 1–100 | personal duels, vanguard charges |
| governance | int 1–100 | city yield, trade efficiency |
| integrity | int 1–100 | loyalty resilience, bribery immunity |
| loyalty | int 1–100 | bond to current lord; drifts each turn based on salary, battles, essence |
| faction | string | currently serving faction; update on defection |
| health | int 1–100 | officer health; declines with age, wounds, illness |
| age | int | current age; officers die of old age or in battle |
| experience | int 0–9999 | accumulated XP; thresholds unlock stat growth |
| city_id | string\|null | assigned city; null = unassigned / in pool |
| army_id | string\|null | attached army; null = not in field |
| tags | list[Tag] | specialist classifiers |

### City

| field | type | notes |
| :--- | :--- | :--- |
| id | string | unique identifier |
| name | string | romanised name |
| chinese | string | Han character name |
| region | string | strategic region |
| terrain | Terrain | terrain type |
| x, y | int | grid position (x: 0–19, y: 0–14) |
| population | int | static for MVP; caps recruitment per turn |
| faction | string | controlling faction |
| governor_id | string\|null | assigned governor officer; applies governance bonus |
| garrison | int | troops stationed for city defence |
| agriculture | int 1–100 | AG pillar — active |
| commerce | int 1–100 | COM pillar — active |
| technology | int 1–100 | TECH pillar — active in yield formula; BUILD command deferred |
| order | int 1–100 | ORD pillar — active in corruption formula; BUILD command deferred |
| defense | int 1–100 | DEF pillar — active |
| food | int | food stockpile in units; army upkeep drawn here |
| gold | int | treasury; recruitment and diplomacy costs drawn here |

### Army

| field | type | notes |
| :--- | :--- | :--- |
| id | string | unique identifier |
| faction | string | owning faction |
| general_id | string | commanding officer (strategy drives formation size, movement) |
| officer_ids | list[string] | attached subordinate officers |
| troops | int | current troop count |
| morale | int 1–100 | battle effectiveness multiplier; drops on retreat/defeat |
| supply | int | turns of food remaining; 0 = attrition |
| x, y | int | current map position |
| stance | Stance | idle / march / siege / encamp |

### Element

Wood, Fire, Earth, Metal, Water

### Terrain

Plain, Mountain, Forest, River, Coast, Pass

### Tag

**classification tags**:

| tag | meaning |
| :--- | :--- |
| civil | civil administrator |
| military | military commander |
| lord | eligible sovereign; generates CP; +20% governance |

**specialist tags**:

| tag | bonus | situation |
| :--- | :--- | :--- |
| naval | +25% strategy | river / sea combat |
| cavalry | +20% valour | open-terrain charge |
| vanguard | +15% valour | duel or charge |
| scholar | +20% governance | academy / civil admin |
| diplomat | +20% integrity | alliance / espionage defence |
| engineer | +20% governance | siege / infrastructure |
| merchant | +25% governance | active trade route |
| strategist | +15% strategy | tactical positioning |
| loyalist | +25% integrity | enemy bribery attempt |
| pioneer | +15% governance | newly annexed territory |

---

## data archives

| archive | count | notes |
| :--- | :--- | :--- |
| `data/officers.yaml` | 498 officers | Wei, Shu, Wu, and independent factions |
| `data/cities.yaml` | 30 cities | fixed grid positions, all five pillars, faction assignment |

target city count is 41–47. current 30 covers the core theatre; southern (Jiao Zhou) and far-north (Liaodong) cities are the main gaps.

archives are the canonical human-readable source. they seed the in-memory ledger on new game start. subsequent launches load from `save.json` instead.

---

## scenario & player setup

before the game loop begins, the player must define the starting conditions of the simulation.

### 1. scenario selection

scenarios define the historical epoch, including the distribution of cities, the status of factions, and the available pool of officers.

| scenario | epoch | description |
| :--- | :--- | :--- |
| Yellow Turban Rebellion | AD 184 | the fall of the Han; Zhang Jiao's uprising |
| Dong Zhuo's Rise | AD 189 | chaos in the capital; the anti-Dong Zhuo coalition forms |
| Rivalry of Lords | AD 194 | the death of Dong Zhuo; independent lords vie for power |
| Battle of Guandu | AD 200 | the clash between Cao Cao and Yuan Shao |
| Three Kingdoms | AD 220 | the formal division of the empire into Wei, Shu, and Wu |

**data mapping:**
- scenario selection filters the loaded JSON archives to set the initial `faction` ownership and officer availability in the in-memory ledger.

### 2. sovereign selection

the player chooses an officer to serve as their **Sovereign (君主)**. 

- only officers with the **lord** tag are eligible for selection.
- the selected officer's `essence` determines the player's elemental alignment.
- the sovereign's `strategy` stat determines the initial base **Command Point (CP)** budget.
- if the sovereign dies without an heir (family or assigned successor), the simulation ends in defeat.

### 3. starting position

the player's starting territory is determined by the faction assigned to their chosen sovereign in the selected scenario.

- **Main Capital**: the primary city where the sovereign is stationed.
- **Territory**: all cities assigned to the same faction in the scenario data.
- **Resources**: starting food and gold are calculated as the sum of controlled cities' outputs.

---

## game systems

---

### 1. strategic map (china campaign map)

the china map is the permanent primary surface — it is never hidden or replaced. all other screens are overlays or panels that open over the map and close back to it.

**screen hierarchy:**
```
StrategicMapScreen (always present)
  ├── CityScreen      — overlay; opens on city click
  ├── OfficerScreen   — overlay; opens from city or officer roster
  ├── ArmyScreen      — overlay; opens on army token click
  ├── DiplomacyScreen — overlay; opens from action menu
  └── BattleScreen    — sub-event; fires on army contact; returns to map on close
```

the map shows all cities, armies, terrain, and faction territories. all player commands (city builds, officer assignments, army orders, diplomacy) are issued from or through screens that layer on top of it.

**map grid:**
- canvas: 20 × 15 tiles (x west→east, y south→north).
- each tile is one terrain type; city tiles overlay terrain.
- target: 41–47 cities positioned to reflect historical geography.

**geographic regions (九州 nine provinces):**

| region | 州 | rough tile zone | key cities |
| :--- | :--- | :--- | :--- |
| Si Li (capital) | 司隸 | x 6–11, y 9–11 | Luoyang, Chang'an, Tongguan |
| Ji Zhou (Hebei) | 冀州 | x 8–13, y 11–14 | Ye, Beiping |
| You Zhou (north-east) | 幽州 | x 12–17, y 12–14 | Youzhou, Liaodong |
| Bing Zhou (north-west) | 并州 | x 5–9, y 11–13 | Jinyang |
| Liang Zhou (far west) | 涼州 | x 1–6, y 8–12 | Tianshui, Wuwei, Dunhuang |
| Yu Zhou / Xu Zhou (central-east) | 豫州 / 徐州 | x 10–16, y 7–10 | Xuchang, Xuzhou, Xiapi |
| Yang Zhou (south-east) | 揚州 | x 11–17, y 4–8 | Jianye, Hefei, Kuaiji |
| Jing Zhou (central-south) | 荊州 | x 7–12, y 3–8 | Xiangyang, Jiangling, Changsha |
| Yi Zhou (south-west / Shu) | 益州 | x 2–8, y 2–8 | Chengdu, Hanzhong, Jiamenguan |
| Jiao Zhou (far south) | 交州 | x 6–12, y 0–3 | Nanhai, Jianning (planned) |

**map layers (Godot TileMap):**

| layer | content |
| :--- | :--- |
| terrain | base tile per cell: plain, mountain, forest, river, coast, pass |
| geography | fixed overlays: Yellow River, Yangtze, Qinling range, Taihang range |
| faction territory | coloured region fill per controlling faction; updates on city capture |
| city nodes | icon at city tile: size = population tier (small/medium/large); colour = faction |
| army tokens | unit marker at army position; number = troop count tier |
| season overlay | tint shift per season: green (spring), amber (summer), gold (autumn), grey (winter) |
| ui overlay | city name labels, army faction flags, selected unit highlight |

**city node tiers (population):**

| tier | population | visual |
| :--- | :--- | :--- |
| village | < 100k | small dot |
| town | 100k–250k | medium icon |
| city | 250k–400k | large icon |
| capital | > 400k | large icon + wall ring |

**navigation:**
- camera pans with arrow keys or drag; zooms with scroll wheel.
- clicking a city node → opens CityScreen overlay.
- clicking an army token → opens ArmyScreen overlay.
- two opposing armies on adjacent tiles → march order triggers battle on cycle C settlement.
- end-turn button advances all cycles and re-renders map state.

**movement:**
- `march <army_id> <x> <y>` in cycle C queues movement.
- movement range per turn: `floor(general.strategy ÷ 20)`, min 1, max 5 tiles.
- terrain movement costs (tiles per step):

| terrain | land army | naval army |
| :--- | :---: | :---: |
| plain | 1 | — |
| forest | 2 | — |
| river | 2 | 1 |
| coast | 2 | 1 |
| mountain | 3 | — |
| pass | 2 | — |

- naval armies (general has naval tag) move freely on river/coast; 3× cost on land tiles.
- army entering a tile with an enemy army triggers battle (cycle D).
- army adjacent to enemy city with siege stance: DEF degrades each turn.

**supply line:**
- supply check = Chebyshev distance to nearest friendly city ≤ 5 tiles.
- if out of range: supply countdown begins (5 turns before attrition). no pathfinding required.

**data requirements:**
- `armies` table: faction, general_id, troops, morale, supply, x, y, stance.
- `map_tiles` table: x, y, terrain (for non-city tiles beyond city list).

**status:** city grid and terrain types exist in archive. army model, map_tiles table, Godot TileMap rendering, and movement not yet implemented.

---

### 2. city development

each city is an economic unit that generates food, gold, and manpower each turn. the player allocates commands to grow pillars and sustain their state.

**active pillars (MVP — 3 pillars):**

| pillar | code | produces | seasonal peak |
| :--- | :--- | :--- | :--- |
| agriculture | AG | food per turn | autumn |
| commerce | COM | gold per turn | summer |
| defense | DEF | wall strength; reduces siege damage | — |

TECH and ORD are read from city archive data and applied in yield formulas (TECH amplifies AG/COM output; ORD reduces corruption loss). player BUILD commands for these pillars are deferred — starting values are set by the scenario data only.

**derived outputs per turn:**
- `food += (AG × season delta) − army_upkeep`
- `gold += (COM × season delta) − officer_salaries`
- population is static for MVP; no growth/decline simulation.

**governor bonus:**
- assigning a civil officer as governor applies `governance × 0.1` as a flat bonus to all pillar growth commands in that city.
- engineer tag: +20% to defense build commands.
- scholar tag: +20% to technology build commands.
- merchant tag: +25% to commerce build commands.

**siege state:**
- city under siege: no AG/COM income; DEF degrades each turn by attacker's siege rating.
- if DEF reaches 0: city falls; garrison troops destroyed; faction transfers.

**status:** pillars, seasonal deltas, and CP commands implemented. governor assignment, food/gold resources, population dynamics, unrest, and siege not yet implemented.

---

### 3. officer management

officers are the primary asset of any state. they are recruited, assigned to roles, and retained through loyalty management.

**lifecycle:**
1. **discovery** — officers appear via search actions, events, or rival defection.
2. **recruitment** — player spends gold and CP to make an offer; officer accepts based on loyalty, rival offers, and elemental resonance with sovereign.
3. **assignment** — officer assigned as governor (city), general (army), or advisor (lord's staff).
4. **service** — officer stats drift seasonally; loyalty drifts based on salary, treatment, wins/losses.
5. **exit** — officer retires (age/health), defects (low loyalty + rival offer), dies in battle or duel, or is dismissed.

**loyalty drift (per turn):**
- base decay: −1 per turn if unpaid.
- salary paid: +1 per turn.
- winning battles: +2.
- defeat or heavy losses: −3.
- elemental resonance with sovereign's essence: +1 to +3 per turn.
- rival bribery attempt: −5 to −20 (resisted by integrity).
- if loyalty < 20: officer may defect next turn.

**assignment roles:**

| role | effect |
| :--- | :--- |
| governor | applies governance bonus to city pillars; manages food/gold output |
| general | commands an army; strategy drives CP allowance for battle commands |
| advisor | attached to lord; provides passive bonus to CP budget |
| unassigned | in the officer pool; earns salary but provides no bonus |

**experience (XP):**
- officers gain XP through active service; unassigned officers gain nothing.
- XP sources:

| action | XP gained |
| :--- | :--- |
| battle won (as general) | +200 |
| battle won (as subordinate) | +80 |
| duel won | +150 |
| governing a city for 1 year (12 turns) | +50 |
| successful search / diplomacy action | +30 |
| city developed under governance | +10 per pillar raised |

- XP thresholds unlock +1 to a relevant stat (strategy for generals, governance for governors) and may trigger a rank promotion event in the ledger.

| threshold | tier | reward |
| :--- | :--- | :--- |
| 500 | novice | +1 to primary stat |
| 1500 | veteran | +2 to primary stat |
| 3500 | master | +3 to primary stat; new tag eligible |

**search mechanic:**
- `search <city_id>` command: costs 1 CP; chance to discover an unrecruited officer in the region.
- search success rate scales with governance of the assigned officer and city's COM pillar.
- found officers appear in the ledger_log as OFFICER_DISCOVERED events.

**data:** loyalty and faction live directly on the Officer record. defection = update `loyalty`, `faction`, `city_id`, `army_id` in place; log the event to `ledger_log`.

**status:** officer archive, stats, essence drift, and tags implemented. assignment, loyalty drift, search, age/health, and defection not yet implemented.

---

### 4. army system

armies are the instrument of territorial expansion and defence. they are raised, trained, moved, and spent.

**army composition:**
- each army has one commanding general and up to 3 subordinate officers.
- general's strategy stat determines max troop cap for the army formation.
- subordinate tags apply situational bonuses during battle (naval, cavalry, engineer, etc.).

**troop recruitment:**
- `recruit <city_id> <amount>` command: costs gold and 1 CP; draws from city population.
- max recruitable per turn: `population × 0.05`.
- trained troops cost more but have higher base morale.

**supply and logistics:**
- each army consumes food per turn: `troops ÷ 100` units from the nearest friendly city.
- if supply drops to 0: morale −5 per turn; troops −2% per turn (attrition).
- supply line: Chebyshev distance to nearest friendly city ≤ 5 tiles; out of range = countdown begins.

**army stances:**

| stance | effect |
| :--- | :--- |
| idle | stationed; no movement; full resupply |
| march | moving toward target tile |
| siege | attacking a city; DEF degrades each turn |
| encamp | fortified position; −50% incoming damage; no advance |

**movement:**
- armies move via player orders in cycle C: `march <army_id> <x> <y>`.
- movement range per turn: `general.strategy ÷ 20`, min 1, max 5 tiles.
- entering a tile occupied by an enemy army triggers battle resolution in cycle C.

**status:** army model not yet implemented. troop count tracked only as garrison on cities.

---

### 5. tactical battle map (Auto-Resolve)

initially implemented as a mathematical "auto-resolve" system. battle is a sub-event within the strategic turn — it fires during cycle C settlement, resolves fully, and writes results back to the ledger before the month closes.

**time scale:**
- strategic time: 1 turn = 1 month.
- battle time: 1 round = 1 day. a typical engagement runs 3–7 rounds within the same month.
- battle does not consume a separate turn — it happens inside the month in which armies make contact.

**trigger conditions:**
- field battle: army moves onto tile occupied by enemy army.
- sortie: defender orders a sortie while city is under siege.

**battle resolution per round:**

```
damage = (attacker.troops × valour_mod × terrain_mod × element_mod × morale_mod) ÷ 100
defender.troops -= damage
defender.morale -= damage ÷ 50
```

modifiers:
- `valour_mod` = attacking officer's valour ÷ 50 (range 0.02–2.0)
- `terrain_mod` = base terrain multiplier (plain: 1.0, mountain: 0.8, etc.)
- `element_mod` = essence drift multiplier of commanding officer
- `morale_mod` = attacker morale ÷ 100

**duel mechanics:**
- resolution: both generals roll `valour + 1d20`; higher roll wins.
- decisive win (diff > 15): loser dies or flees; −30 morale to loser's army.

**status:** battle system not yet implemented. auto-resolve formula is the initial implementation target.

---

### 6. diplomacy

diplomatic actions are issued in cycle B alongside other player commands. gold cost instead of CP.

**relation score:**
- each pair of factions has a relation score (−100 hostile to +100 allied).

**diplomatic actions (cost CP or gold):**
- send tribute, propose alliance, threaten.

**status:** not yet implemented.

---

### 7. events system

cycle A generates world events each turn. initially limited to seasonal shifts and resource signals.

**status:** seasonal events implemented. random and historical events deferred.

---

### 8. turn structure (full)

each turn = one month. three cycles execute in strict order.

```
┌─────────────────────────────────────────────────────────┐
│  cycle A — world update (automatic, no player input)    │
│    1. bazi clock advances                               │
│    2. seasonal deltas calculated for all cities         │
│    3. essence drift recalculated for all officers       │
│    4. random / historical events generated              │
│    5. loyalty drift calculated                          │
│    6. army supply / attrition computed                  │
├─────────────────────────────────────────────────────────┤
│  cycle B — player commands (CP-limited)                 │
│    - city development: ag / com / def                   │
│    - officer: search / recruit / assign / dismiss       │
│    - military: recruit / march / siege / encamp         │
│    - diplomacy: tribute / alliance / threaten (gold)    │
│    - commands queued; engine validates CP budget        │
├─────────────────────────────────────────────────────────┤
│  cycle C — settlement                                   │
│    - apply all cycle A / B deltas sequentially         │
│    - resolve battles (sub-event; 3–7 day rounds each)  │
│    - resolve siege outcomes                             │
│    - advance game_clock                                 │
│    - append all events to ledger_log                    │
└─────────────────────────────────────────────────────────┘
```

---

### 9. essence drift

effective stat = clamp(base × drift, 1, 100).

---

### 10. bazi calendar (天干地支)

tracks time using the traditional 60-unit stem-branch cycle. epoch: AD 184.

---

### 11. victory

**win condition:** one faction holds all cities → sovereignty victory. checked at end of each cycle C.

**defeat condition:** sovereign dies with no assigned successor, or all cities lost.

---

## persistence

### in-memory ledger + JSON save

runtime state lives entirely in memory. on save, the ledger serialises to `save.json`. on load, `save.json` is deserialised back into the ledger. new game starts seed from `godot/data/*.json` (converted from YAML archives).

this approach is cross-platform by default — no native extensions required, works on desktop, mobile, and web exports.

| ledger key | type | purpose |
| :--- | :--- | :--- |
| `officers` | Dictionary | officer records: stats, essence, health, age, experience, tags |
| `cities` | Dictionary | city records: pillars, governor_id, garrison, food, gold |
| `armies` | Dictionary | army records: general, troops, morale, supply, position, stance |
| `faction_relations` | Dictionary | pairwise relation scores between factions |
| `game_clock` | Dictionary | current year, month |
| `resources` | Dictionary | player-faction grain and gold totals |
| `logs` | Array[Dictionary] | append-only event log: year, month, cycle, event_type, description, effects |

**settlement:** cycle C applies deltas sequentially. no snapshot/restore — keep it simple and fix bugs at the source.

---

## frontend

### scenes

| scene | maps to system | status |
| :--- | :--- | :--- |
| SplashScreen | — | hello world scaffold done |
| ScenarioSelectScreen | scenario & sovereign setup | [ ] |
| StrategicMapScreen | strategic map — tilemap, armies, faction overlays | [ ] |
| CityScreen | city development — pillar bars, officer slot, food/gold | [ ] |
| OfficerScreen | officer management — roster, stats, assignment | [ ] |
| ArmyScreen | army system — troop count, morale, supply, stance | [ ] |
| BattleScreen | tactical battle — report / auto-resolve animation | [ ] |
| DiplomacyScreen | diplomacy — faction list, relation scores, action panel | [ ] |
| LedgerScreen | ledger log — scrollable history, event filter | [ ] |
| VictoryScreen | victory — unification win / defeat screen | [ ] |

---

## headless testing strategy

since the engine is decoupled from the frontend, verification is performed through automated logic tests and simulation snapshots.

### 1. the "robot player" (integration test)

a standalone script (GDScript or Python) that interacts with the `headless` modules to simulate gameplay. this is the primary tool for Milestone 1 verification.

**structure:**
```text
1. setup:
   - instantiate ledger and call load_data() to seed from JSON archives.
   - instantiate clock at AD 189, month 1.
   - instantiate engine with ledger, clock, and sovereign_id.

2. execution (simulate 12 turns):
   - for turn in 1..12:
     - log state: current year/month/dominant_element.
     - cycle A: run world update (essence drift, resource deltas, loyalty drift).
     - cycle B: inject player commands (e.g., build_ag, recruit_army, send_tribute).
     - cycle C: run settlement (apply deltas sequentially).

3. validation:
   - assert: year/month progressed correctly.
   - assert: city pillars changed by expected seasonal + command deltas.
   - assert: officer stats were correctly clamped (1–100) after drift.
   - check: 'ledger_log' contains a record for every event in the turn.
```

### 2. unit testing (logic isolation)

each headless module must pass isolated tests:
- **clock_test**: verify the 60-unit cycle and seasonal transitions.
- **essence_test**: verify that every Wu Xing relationship (nourishing, controlling, etc.) produces the correct drift multiplier.
- **battle_math_test**: run 10,000 auto-resolve loops to verify that stat differences result in consistent, non-glitchy outcomes (no negative troop counts, etc.).

### 3. "golden master" regression

- run a 100-turn simulation with a fixed random seed.
- serialise the final ledger state to a text file (`master_state.json`).
- after any engine refactor, re-run the simulation. 
- if the output differs from the master state, a logic regression has occurred.

---

## milestones

> `done` · `in progress` · `not started`

### development strategy (simplifications)

to manage complexity and reach a "playable" state faster, the following strategies are enforced:
- **auto-resolve first**: milestone 6 (battle) will initially implement a side-effect-free math resolver (valour vs strategy) rather than a full tactical grid. the grid is deferred to expansions.
- **canonical scenario**: focus exclusively on **AD 189 (dong zhuo's rise)** for initial development and balancing. other scenarios are data-only targets for now.
- **vertical slice seeding**: start with a subset of the 500 officers (e.g., only those with the `lord` tag and their primary generals) to verify essence drift and the game loop.
- **ui minimalism**: use placeholder godot ui components to verify engine logic before committing to custom pixel art assets.

### screen development sequence

build screens in this order — each is a concrete "done" test for the engine systems behind it. the china map is always screen 1 and never leaves.

| order | screen | milestone | what it proves |
| :---: | :--- | :---: | :--- |
| 1 | StrategicMapScreen | 2 ✓ | map renders; cities visible; camera and faction colours work |
| 2 | CityScreen | 3 ✓ | pillar data reads from ledger; build commands fire correctly |
| 3 | OfficerScreen | 4 | assignment, loyalty, and XP visible; assign/dismiss commands work |
| 4 | ArmyScreen | 5 | army data visible; movement orders queue and execute |
| 5 | BattleScreen | 6 | auto-resolve fires on army contact; result displayed; ledger updated |
| 6 | DiplomacyScreen | 7 | relation scores visible; tribute and alliance commands fire |
| 7 | ScenarioSelectScreen | 8 | scenario + sovereign selection loads correct ledger state |
| 8 | VictoryScreen | 8 | win / defeat state detects and renders correctly |
| 9 | LedgerScreen | 8 | event log is scrollable; filter works |

### development phases

| phase | focus | milestones | status |
| :--- | :--- | :--- | :--- |
| **1: genesis** | foundation, data pipeline, and turn engine | 0, 1 | [x] |
| **2: cartography** | strategic map and visual world | 2 | [x] |
| **3: governance** | city development and economics | 3 | [x] |
| **4: sovereignty** | officer management and allegiance | 4 | [ ] |
| **5: conflict** | armies and tactical battle system | 5, 6 | [ ] |
| **6: statecraft** | diplomacy and world events | 7 | [ ] |
| **7: legacy** | victory, scoring, and polish | 8 | [ ] |

### milestone tracking

| milestone | focus | status |
| :--- | :--- | :--- |
| 0 | foundation | [x] |
| 1 | data + turn engine (focus: AD 189) | [x] |
| 2 | strategic map (visualizing china) | [x] |
| 3 | cities (development & economics) | [x] |
| 4 | officers (management & loyalty) | [ ] |
| 5 | army system (raising & movement) | [ ] |
| 6 | tactical battle (auto-resolve math) | [ ] |
| 7 | diplomacy + events | [ ] |
| 8 | victory + polish | [ ] |

---

### milestone 0 — foundation `done`

| item | status |
| :--- | :--- |
| Godot 4 project (1280×720, pixel settings, Brewfile, Makefile) | [x] |
| splash screen (fade, blink prompt, 5s auto-advance, key dismiss) | [x] |
| main scene placeholder | [x] |
| design: officer + city archives (YAML, 498 officers, 30 cities) | [x] |
| design: bazi calendar, turn structure, ledger schema | [x] |

---

### milestone 1 — data + turn engine `done`

JSON archive loading, in-memory ledger init, bazi clock, and the turn loop in GDScript. architecture since simplified to 3 cycles. **priority focus: AD 189 scenario.**

| item | status |
| :--- | :--- |
| JSON archive loading + ledger seed | [x] |
| Scenario & Sovereign selection logic | [x] |
| bazi calendar + essence drift | [x] |
| cycle A — seasonal deltas | [x] |
| cycle B — diplomacy | [x] |
| cycle C — player commands | [x] |
| cycle D — atomic settlement | [x] |

---

### milestone 2 — strategic map (cartography) `done`

| item | status |
| :--- | :--- |
| TileMap layers (terrain, geography, faction territory, city nodes, UI overlay) | [x] |
| high-fidelity China polygon map with rivers and islands | [x] |
| city nodes with tier-based icons | [x] |
| city labels hidden by default; revealed on hover | [x] |
| camera pan and zoom controls | [x] |
| faction territory colour fill | [x] |

---

### milestone 3 — city development & economics `done`

| item | status |
| :--- | :--- |
| city pillar model (AG / COM / TECH / ORD / DEF, 1–100) | [x] |
| seasonal delta calculation per pillar | [x] |
| TECH multiplier on AG / COM yield | [x] |
| CP command validation (build pillar) | [x] |
| CityScreen overlay — pillar bars display | [x] |
| food and gold resource model (structure defined) | [x] |

---

### milestone 4 — officer management & allegiance `not started`

| item | tags | status |
| :--- | :--- | :--- |
| officer assignment (governor / general / advisor roles) | `[engine]` `[medium]` | [ ] |
| loyalty drift per turn (salary, battles, essence resonance) | `[engine]` `[medium]` | [ ] |
| rival bribery resistance (integrity check) | `[engine]` `[easy]` | [ ] |
| officer search mechanic (`search <city_id>`) | `[engine]` `[easy]` | [ ] |
| XP system — 3 tiers (novice / veteran / master) | `[engine]` `[easy]` | [ ] |
| age and health degradation; officer death | `[engine]` `[medium]` | [ ] |
| defection — update loyalty, faction, city_id, army_id in place; log event | `[engine]` `[medium]` | [ ] |
| OfficerScreen — roster, stats, assignment panel | `[ui]` `[medium]` | [ ] |

---

### milestone 5 — army system `not started`

| item | tags | status |
| :--- | :--- | :--- |
| army data model (general, troops, morale, supply, stance) | `[engine]` `[medium]` | [ ] |
| troop recruitment command | `[engine]` `[easy]` | [ ] |
| supply logistics calculation (food draw per turn) | `[engine]` `[medium]` | [ ] |
| army movement range calculation | `[engine]` `[medium]` | [ ] |
| supply check (Chebyshev distance ≤ 5 to nearest friendly city) | `[engine]` `[easy]` | [ ] |
| attrition when supply cut | `[engine]` `[easy]` | [ ] |
| StrategicMapScreen — army tokens and movement orders | `[ui]` `[hard]` | [ ] |
| ArmyScreen — troop count, morale, supply, stance | `[ui]` `[medium]` | [ ] |

---

### milestone 6 — tactical battle (auto-resolve) `not started`

| item | tags | status |
| :--- | :--- | :--- |
| auto-resolve formula (valour vs strategy) | `[engine]` `[medium]` | [ ] |
| casualty and morale calculation | `[engine]` `[easy]` | [ ] |
| duel resolution math | `[engine]` `[easy]` | [ ] |
| outcome report (ledger log) | `[engine]` `[easy]` | [ ] |
| BattleScreen — auto-resolve report display | `[ui]` `[easy]` | [ ] |
| tactical grid (visual) | — | [-] |

---

### milestone 7 — diplomacy & events `not started`

| item | tags | status |
| :--- | :--- | :--- |
| faction relation score table (−100 to +100 per pair) | `[engine]` `[easy]` | [ ] |
| diplomatic actions — tribute, alliance, threaten | `[engine]` `[medium]` | [ ] |
| diplomacy command processing in cycle B | `[engine]` `[medium]` | [ ] |
| random event generator | `[engine]` `[medium]` | [ ] |
| historical event triggers (scripted) | `[engine]` `[hard]` | [ ] |
| DiplomacyScreen — faction list, relation scores, action panel | `[ui]` `[medium]` | [ ] |

---

### milestone 8 — victory & polish `not started`

| item | tags | status |
| :--- | :--- | :--- |
| unification victory check (one faction holds all cities) | `[engine]` `[easy]` | [ ] |
| defeat check (sovereign dead, no successor; or all cities lost) | `[engine]` `[easy]` | [ ] |
| ScenarioSelectScreen — scenario and sovereign selection | `[ui]` `[medium]` | [ ] |
| VictoryScreen — win / defeat screen | `[ui]` `[easy]` | [ ] |
| LedgerScreen — scrollable event history and filter | `[ui]` `[easy]` | [ ] |
| full pixel art asset pass | `[ui]` `[hard]` | [ ] |

## decisions

| decision | choice | why |
| :--- | :--- | :--- |
| in-memory ledger over SQLite | Dictionary + JSON serialisation | cross-platform by default — no GDExtension dependency; SQLite has no built-in Godot 4 support and breaks on web exports; snapshot/restore gives sufficient atomicity for a turn-based sim |
| GDScript over C# | GDScript | tighter Godot 4 integration; no separate build step; sufficient for turn-based simulation pace |
| YAML archives, JSON runtime | YAML → JSON via `make data` | YAML is human-readable and diffable; JSON is fast to load in Godot without a parser dependency |
| headless-first engine | `RefCounted`/`Node` classes testable via `godot --headless` | allows CI verification without the editor; separates simulation correctness from visual rendering |
| auto-resolve battle (v1) | math formula, no tactical grid | reach playable loop sooner; grid deferred to post-release expansion |
| single scenario lock (AD 189) | Dong Zhuo's Rise only | prevents data balancing sprawl before core loop is verified |
| mutable state + append-only log | live tables + `ledger_log` | simplifies reads (no event sourcing reconstruction); full history preserved for victory scoring |
| flatten OfficerAllegiance | loyalty + faction on Officer directly | separate row-lifecycle table is a database pattern, not a game pattern; city_id/army_id already on Officer |
| 3-cycle turn loop | A (world) / B (commands) / C (settle) | diplomacy is just a gold-cost command; 4 cycles added complexity with no player-visible benefit |
| no snapshot/restore | sequential delta apply in cycle C | turn-based game doesn't need transactional rollback; bugs should be fixed not hidden behind recovery logic |
| 3 active pillars (AG/COM/DEF) | TECH and ORD deferred | 5 interacting pillars are hard to balance before the core loop is proven |
| static population | no growth/decline for MVP | population dynamics multiply balancing complexity; static cap on recruitment is sufficient |
| territory-only victory | one faction holds all cities | multi-dimension legacy scoring requires tuning three systems before the game loop exists |
| distance-based supply | Chebyshev ≤ 5 tiles to friendly city | path-finding on faction-owned tile graph is [hard] and a source of edge-case bugs |
| 3-tier XP | 500 / 1500 / 3500 | 5 tiers with differentiated rewards per tier requires playtesting to balance; 3 flat tiers are clear and testable |

---

## complexity score

| dimension | score | notes |
| :--- | :--- | :--- |
| overall | 3 / 5 | moderate; multi-layer simulation with in-memory ledger, headless engine, and Godot frontend |
| core (clock, essence, economy) | 2 / 5 | pure math; stateless functions; well-bounded domain |
| engine (sovereign_engine, ledger) | 3 / 5 | three-cycle turn loop, sequential settlement, CP validation logic |
| data pipeline (YAML → JSON → DB) | 2 / 5 | one-way conversion; `make data` is the only transform step |
| ui (Godot scenes) | 2 / 5 | view-only; reads from ledger; no simulation logic |
| future: army + battle system | 4 / 5 | movement, supply lines, auto-resolve math, and eventual tactical grid |
| future: diplomacy + events | 4 / 5 | pairwise faction state, scripted triggers, branching outcomes |

---

## roadmap

### near term

milestones 4–8; priority order: officers → armies → battle → diplomacy → victory. see milestone tracking above for full task breakdowns.

highest dependency risk items:
- `[engine]` historical event triggers — scripted facts vs generative logic boundary `[hard]`
- `[ui]` StrategicMapScreen army tokens + movement input `[hard]`

### ideas

- `[engine]` **tactical grid battle** — full 15×9 tactical grid; unit movement ranges; detailed terrain modifiers `[hard]`
- `[engine]` **advanced diplomacy & espionage** — spy placement, bribery (resisted by integrity), marriage alliances `[hard]`
- `[engine]` **historical scripted events** — full chronological triggers (Yellow Turban AD 184, Red Cliffs AD 208, etc.) `[hard]`
- `[engine]` **officer lineage & health** — aging stat decay, heir system, complex wound and illness management `[medium]`
- `[engine]` **naval warfare** — specialised naval units; river/sea tactical grids; boarding mechanics `[hard]`
- `[engine]` **multi-era engine** — Sengoku, Roman Republic, and other eras on the same sovereign engine `[hard]`
- `[engine]` **AI sovereigns** — each NPC faction runs an AI decision-maker in cycle B; see design below `[medium]`

---

### ai sovereign design

every faction not controlled by the player is governed by an AI sovereign. the AI runs in cycle B as a pure function — headless-testable, no UI dependency.

**faction ai record (two fields on ledger faction entry):**

| field | type | notes |
| :--- | :--- | :--- |
| `ai_intelligence` | int 1–4 | how well the AI evaluates the game state |
| `ai_behaviour` | string | primary decision-making priority |

**intelligence levels:**

| level | name | description |
| :---: | :--- | :--- |
| 1 | reckless | random command selection; ignores supply and debt; no planning horizon |
| 2 | standard | basic heuristics — build economy when safe, recruit before attacking, target weakest neighbour |
| 3 | shrewd | evaluates city values, maintains supply lines, coordinates army timing across turns |
| 4 | brilliant | optimises essence drift timing, anticipates player moves, defers attack until overwhelming advantage |

**behaviour profiles (priority weights):**

| behaviour | military | economy | diplomacy | character |
| :--- | :---: | :---: | :---: | :--- |
| aggressive | high | low | low | attacks first; strikes before enemy consolidates |
| defensive | low | high | medium | fortifies DEF; avoids offensive wars; counter-attacks only |
| expansionist | high | medium | low | rapid territory grab; spreads resources thin |
| diplomatic | low | medium | high | prefers alliances and tribute; fights only when cornered |
| economic | low | high | medium | maximises AG/COM before raising armies; slow but powerful late |
| opportunistic | medium | medium | low | targets the weakest neighbour each turn; retreats when losing |

**decision function (pseudo-code):**
```
ai_decide(ledger, faction_id, intelligence, behaviour) -> List[Command]:
    state = ledger.snapshot_for(faction_id)
    priorities = behaviour_weights[behaviour]
    candidates = generate_candidate_commands(state, priorities)
    if intelligence == 1: return random_sample(candidates)
    scored = score_commands(candidates, state, intelligence)
    return top_commands_within_cp_budget(scored)
```

**preset ai profiles (AD 189 scenario):**

| faction | intelligence | behaviour |
| :--- | :---: | :--- |
| Dong Zhuo | 3 | aggressive |
| Cao Cao | 4 | opportunistic |
| Yuan Shao | 3 | expansionist |
| Liu Biao | 2 | defensive |
| Sun Jian | 3 | aggressive |
| Liu Zhang | 2 | economic |
| independent lords | 1–2 | defensive |
