# spec: 帝王战纪：三国录

> "Sovereignty through the Ledger, Strategy through the Elements."

## tldr
A headless, turn-based grand strategy engine set in the Three Kingdoms era. Architecture prioritises **data integrity (SQLite)** and **simulation purity (Headless)** over visual spectacle. 

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

## scope containment policy (no-drift)
- **Engine/UI Separation**: Logic MUST be testable in `make test` without initializing the UI.
- **SQLite Ledger**: If a feature doesn't store data in the ledger, it is not part of the core simulation.
- **No Tactical Grid (V1)**: The grid battle is deferred to expansions. Stick to the auto-resolve formula.
- **Single Scenario**: Development is locked to **AD 189 (Dong Zhuo)**. Do not balance other scenarios until the game is playable.
- **No "Just-in-case"**: Do not implement features until the turn engine needs them (e.g., don't build espionage before the diplomacy cycle is functional).

---

## overview
...

**帝王战纪：三国录** is a turn-based grand strategy simulation set during China's Three Kingdoms era (AD 184–280), beginning at the Yellow Turban Rebellion epoch. the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft, elemental alignment, and long-term institutional building.

victory is measured across three dimensions — territory, institutional strength, and historical legacy — all chronicled in the ledger.

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
- **mutable current state + chronicle log**: SQLite tables store live game state directly. `ledger_log` is an append-only chronicle for history, narrative, and victory scoring — not the primary state store.
- **YAML archives**: human-readable canonical data (`data/officers.yaml`, `data/cities.yaml`). read once at first-run seed; never touched at runtime.

### module responsibilities

| module | role |
| :--- | :--- |
| main | entry point: seeds DB from YAML archives on first run, launches frontend |
| models | typed domain models: Officer, City, Army, Element, Tag, Terrain, and bonus lookup tables |
| clock | bazi calendar: heavenly stem (天干) / earthly branch (地支) tracking, essence drift calculation |
| engine | three-cycle turn processor: command validation, seasonal deltas, CP budget |
| battle | tactical battle resolver: grid movement, combat, duels, siege, morale |
| diplomacy | inter-faction relations: alliances, espionage, tribute, bribery |
| events | event generator: random events, historical triggers, NPC appearances |
| ledger | SQLite layer: schema init, CRUD, atomic settle_turn transaction |
| archive | YAML loader: parses human-readable config into validated domain models |
| frontend | Godot 4 scenes: splash, strategic map, city, officer, army, battle, diplomacy, ledger, victory |

### data flow

```
data/officers.yaml ──┐
data/cities.yaml     ├──> archive ──> domain models
                     │
                     └──> main (first-run seed) ──> ledger ──> ledger.db
                                                                    ↑
    frontend ──> engine (cycle A/B/C/D) ─────────────> ledger ──────┘
             └──> battle / diplomacy / events ────────────────────┘
                                                  (settle_turn, atomic commit)
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
| health | int 1–100 | officer health; declines with age, wounds, illness |
| age | int | current age; officers die of old age or in battle |
| experience | int 0–9999 | accumulated XP; thresholds unlock stat growth and rank promotion |
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
| population | int | current population; grows/shrinks based on AG and ORD |
| faction | string | controlling faction |
| governor_id | string\|null | assigned governor officer; applies governance bonus |
| garrison | int | troops stationed for city defence |
| agriculture | int 1–100 | AG pillar |
| commerce | int 1–100 | COM pillar |
| technology | int 1–100 | TECH pillar |
| order | int 1–100 | ORD pillar |
| defense | int 1–100 | DEF pillar (walls, fortifications) |
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

### OfficerAllegiance

relational state between an officer and a faction. one active row per officer at any time; closed (end_turn set) on defection, dismissal, or death.

| field | type | notes |
| :--- | :--- | :--- |
| officer_id | string | FK → officers |
| faction | string | faction the officer currently serves |
| lord_id | string | the specific lord officer (faction head) |
| loyalty | int 1–100 | bond strength to current lord; drifts per turn |
| joined_turn | int | turn number when service began |
| end_turn | int\|null | turn number when service ended; null = active |

intrinsic officer stats (strategy, valour, etc.) are unchanged by allegiance. when an officer defects, the old row is closed and a new row inserted — full service history preserved in the ledger.

---

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

archives are the canonical human-readable source. they seed the SQLite ledger on first run only. subsequent launches read exclusively from `ledger.db`.

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
- scenario selection filters the `data/cities.yaml` and `data/officers.yaml` to set the initial `faction` ownership and officer availability in `ledger.db`.

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

the primary game view. a stylised top-down map of Han China showing all cities, armies, terrain, and faction territories. the player navigates, issues commands, and transitions to city or battle screens from here.

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
- two opposing armies on adjacent tiles → march order triggers battle on cycle D settlement.
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
- supply line = path of friendly tiles back to nearest friendly city.
- if supply line broken by enemy army: supply countdown begins (default 5 turns before attrition).

**data requirements:**
- `armies` table: faction, general_id, troops, morale, supply, x, y, stance.
- `map_tiles` table: x, y, terrain (for non-city tiles beyond city list).

**status:** city grid and terrain types exist in archive. army model, map_tiles table, Godot TileMap rendering, and movement not yet implemented.

---

### 2. city development

each city is an economic unit that generates food, gold, and manpower each turn. the player allocates commands to grow pillars and sustain their state.

**pillars (1–100):**

| pillar | code | produces | seasonal peak |
| :--- | :--- | :--- | :--- |
| agriculture | AG | food per turn | autumn |
| commerce | COM | gold per turn | summer |
| technology | TECH | multiplier on AG/COM yield | — |
| order | ORD | reduces corruption; sustains loyalty | transitions |
| defense | DEF | wall strength; reduces siege damage | — |

**derived outputs per turn:**
- `food += (AG × TECH multiplier × season delta) − army_upkeep`
- `gold += (COM × TECH multiplier × season delta) − officer_salaries`
- population grows when AG > 60 and ORD > 50; declines under famine or siege.

**governor bonus:**
- assigning a civil officer as governor applies `governance × 0.1` as a flat bonus to all pillar growth commands in that city.
- engineer tag: +20% to defense build commands.
- scholar tag: +20% to technology build commands.
- merchant tag: +25% to commerce build commands.

**unrest and corruption:**
- if ORD < 30: risk of riot event each turn (reduces AG and COM by 1d10).
- if ORD < 15: rebellion — city may defect to a rival faction or become independent.
- high taxation (COM-heavy policy) suppresses ORD over time.

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

| threshold | reward |
| :--- | :--- |
| 100 | +1 to primary stat |
| 300 | +1 to primary stat; rank promotion eligible |
| 700 | +2 to primary stat; new tag eligible |
| 1500 | +2 to primary + secondary stat |
| 3000 | +3 across primary, secondary, tertiary stats |

**search mechanic:**
- `search <city_id>` command: costs 1 CP; chance to discover an unrecruited officer in the region.
- search success rate scales with governance of the assigned officer and city's COM pillar.
- found officers appear in the ledger_log as OFFICER_DISCOVERED events.

**data requirements:**
- `officer_allegiance` table: loyalty drift written here during cycle A; new row inserted on defection.
- `officer_assignments` table: officer_id, role (governor/general/advisor), city_id or army_id, assigned_turn.

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
- supply lines are cut if enemy armies are between the army and its home city.

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
- entering a tile occupied by an enemy army triggers battle resolution in cycle D.

**status:** army model not yet implemented. troop count tracked only as garrison on cities.

---

### 5. tactical battle map (Auto-Resolve)

initially implemented as a mathematical "auto-resolve" system. the battle plays out in the engine before returning the outcome to the strategic map.

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

inter-faction relations managed through a dedicated diplomacy cycle resolved between cycle A and cycle C.

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

each turn = one month. cycles execute in strict order; nothing is committed until cycle D completes.

```
┌─────────────────────────────────────────────────────────┐
│  cycle A — world update (automatic, no player input)    │
│    1. bazi clock advances                               │
│    2. seasonal deltas calculated for all cities         │
│    3. essence drift recalculated for all officers       │
│    4. random / historical events generated              │
│    5. NPC actor events offered                          │
│    6. loyalty drift calculated                          │
│    7. army supply / attrition computed                  │
│    8. population growth/decline computed                │
├─────────────────────────────────────────────────────────┤
│  cycle B — diplomacy (optional, player-initiated)       │
│    - send tribute, propose alliance, bribe, spy         │
│    - no CP cost; costs gold or dedicated diplomat CP    │
├─────────────────────────────────────────────────────────┤
│  cycle C — player commands (CP-limited)                 │
│    - city development: ag / com / tech / ord / def      │
│    - officer: search / recruit / assign / dismiss       │
│    - military: recruit / march / siege / encamp         │
│    - commands queued; engine validates CP budget        │
├─────────────────────────────────────────────────────────┤
│  cycle D — settlement (atomic SQLite transaction)       │
│    - apply all cycle A / B / C deltas                   │
│    - resolve battles triggered by army movement         │
│    - resolve siege outcomes                             │
│    - advance game_clock                                 │
│    - append all events to ledger_log                    │
│    - rollback on failure; ledger unchanged              │
└─────────────────────────────────────────────────────────┘
```

---

### 9. essence drift

effective stat = clamp(base × drift, 1, 100).

---

### 10. bazi calendar (天干地支)

tracks time using the traditional 60-unit stem-branch cycle. epoch: AD 184.

---

### 11. victory and legacy

victory is assessed from the ledger at three levels.

**era end trigger:**
- unification: one faction holds all cities → sovereignty victory.
- year 280 (era end): game scores all dimensions.

---

## persistence

### SQLite (`ledger.db`)

WAL mode and foreign keys enabled.

| table | purpose |
| :--- | :--- |
| `officers` | intrinsic officer records (stats, essence, health, age, experience) |
| `officer_tags` | junction table: officer_id → tag |
| `officer_allegiance` | relational state: faction, lord_id, loyalty, joined_turn, end_turn |
| `officer_assignments` | role, city_id or army_id, assigned_turn |
| `cities` | city records with all pillar values, governor_id, garrison, food, gold |
| `armies` | army records: general, troops, morale, supply, position, stance |
| `army_officers` | junction: army_id → officer_id |
| `faction_relations` | pairwise relation scores between factions |
| `game_clock` | single row: current year, month |
| `ledger_log` | append-only event log (year, month, cycle, event_type, description, effects_json) |

---

## frontend

### scenes

| scene | maps to system | status |
| :--- | :--- | :--- |
| SplashScreen | — | hello world scaffold done |
| ScenarioSelectScreen | scenario & sovereign setup | not started |
| StrategicMapScreen | strategic map — tilemap, armies, faction overlays | not started |
| CityScreen | city development — pillar bars, officer slot, food/gold | not started |
| OfficerScreen | officer management — roster, stats, assignment | not started |
| ArmyScreen | army system — troop count, morale, supply, stance | not started |
| BattleScreen | tactical battle — report / auto-resolve animation | not started |
| DiplomacyScreen | diplomacy — faction list, relation scores, action panel | not started |
| LedgerScreen | ledger log — scrollable history, event filter | not started |
| VictoryScreen | legacy scoring — three-dimension breakdown | not started |

---

## headless testing strategy

since the engine is decoupled from the frontend, verification is performed through automated logic tests and simulation snapshots.

### 1. the "robot player" (integration test)

a standalone script (GDScript or Python) that interacts with the `headless` modules to simulate gameplay. this is the primary tool for Milestone 1 verification.

**structure:**
```text
1. setup:
   - create an in-memory or temporary sqlite database.
   - run 'archive' module to load YAML data.
   - run 'ledger' init to seed the database for AD 189.

2. execution (simulate 12 turns):
   - for turn in 1..12:
     - log state: current year/month/dominant_element.
     - cycle A: run world update (essence drift, resource deltas).
     - cycle B: (optional) inject a diplomacy command.
     - cycle C: inject player commands (e.g., build_ag, recruit_army).
     - cycle D: run settlement (commit to DB, resolve battles).

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
- dump the final state of all SQLite tables to a text file (`master_state.txt`).
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

### development phases

| phase | focus | milestones | status |
| :--- | :--- | :--- | :--- |
| **1: genesis** | foundation, sqlite, and turn engine | 0, 1 | in progress |
| **2: cartography** | strategic map and visual world | 2 | not started |
| **3: governance** | city development and economics | 3 | not started |
| **4: sovereignty** | officer management and allegiance | 4 | not started |
| **5: conflict** | armies and tactical battle system | 5, 6 | not started |
| **6: statecraft** | diplomacy and world events | 7 | not started |
| **7: legacy** | victory, scoring, and polish | 8 | not started |

### milestone tracking

| milestone | focus | status |
| :--- | :--- | :--- |
| 0 | foundation | done |
| 1 | data + turn engine (focus: AD 189) | not started |
| 2 | strategic map (visualizing china) | done |
| 3 | cities (development & economics) | not started |
| 4 | officers (management & loyalty) | not started |
| 5 | army system (raising & movement) | not started |
| 6 | tactical battle (auto-resolve math) | not started |
| 7 | diplomacy + events | not started |
| 8 | victory + polish | not started |

---

### milestone 0 — foundation `done`

| item | status |
| :--- | :--- |
| Godot 4 project (1280×720, pixel settings, Brewfile, Makefile) | done |
| splash screen (fade, blink prompt, 5s auto-advance, key dismiss) | done |
| main scene placeholder | done |
| design: officer + city archives (YAML, 498 officers, 30 cities) | done |
| design: bazi calendar, turn structure, SQLite schema | done |

---

### milestone 1 — data + turn engine `not started`

SQLite init, YAML seeding, bazi clock, and the four-cycle turn loop in GDScript. **priority focus: AD 189 scenario.**

| item | status |
| :--- | :--- | :--- |
| SQLite init (schema, WAL, seed from YAML) | not started |
| Scenario & Sovereign selection logic | not started |
| bazi calendar + essence drift | not started |
| cycle A — seasonal deltas | not started |
| cycle B — diplomacy | not started |
| cycle C — player commands | not started |
| cycle D — atomic settlement | not started |

---

### milestone 6 — tactical battle `not started`

initially implemented as a mathematical "auto-resolve" system to verify army and officer stats before building the visual grid.

| item | status |
| :--- | :--- |
| auto-resolve formula (valour vs strategy) | not started |
| casualty & morale calculation | not started |
| duel resolution math | not started |
| outcome report (ledger log) | not started |
| tactical grid (visual) | deferred |

---

## future ideas & expansions

these systems are deferred to post-release or major expansion phases to maintain focus on the core "Vertical Slice" prototype.

### 1. tactical grid battle
- transition from abstract "auto-resolve" to a full 15×9 tactical grid.
- unit-specific movement and attack ranges.
- detailed terrain modifiers for tactical tiles.

### 2. advanced diplomacy & espionage
- **spy placement**: deep infiltration of enemy cities to sabotage or extract data.
- **bribery**: mechanics for turning high-integrity officers (resisted by integrity).
- **marriage alliances**: familial links between factions for long-term stability.

### 3. historical scripted events
- full chronological event triggers (e.g., Yellow Turban Rebellion AD 184, Red Cliffs AD 208).
- scripted army spawns and territory shifts based on historical timelines.

### 4. officer lifecycle (lineage & health)
- **aging**: officers grow old and stats naturally decay or shift.
- **heirs**: lineage systems to handle faction leadership transitions upon death.
- **health**: complex wound and illness management.

### 5. naval warfare
- specialized naval units and river/sea tactical grids.
- ship-to-ship combat and boarding mechanics.

### 6. multi-era engine
- support for other historical eras (Sengoku, Roman Republic, etc.) using the same core sovereign engine.
