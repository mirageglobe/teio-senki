# spec: 帝王战纪：三国录

> design and architecture reference for the sovereign record engine.

---

## overview

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
| engine | three-phase turn processor: command validation, seasonal deltas, CP budget |
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
    frontend ──> engine (phase A/B/C) ──────────────> ledger ──────┘
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
- two opposing armies on adjacent tiles → march order triggers battle on phase C settlement.
- end-turn button advances all phases and re-renders map state.

**movement:**
- `march <army_id> <x> <y>` in phase B queues movement.
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
- army entering a tile with an enemy army triggers battle (phase C).
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
- `officer_allegiance` table: loyalty drift written here during phase A; new row inserted on defection.
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
- armies move via player orders in phase B: `march <army_id> <x> <y>`.
- movement range per turn: `general.strategy ÷ 20`, min 1, max 5 tiles.
- entering a tile occupied by an enemy army triggers battle resolution in phase C.

**status:** army model not yet implemented. troop count tracked only as garrison on cities.

---

### 5. tactical battle map

a separate scene from the strategic map. triggered when armies clash on the strategic map or when a field sortie is ordered from a besieged city. the battle plays out on a dedicated tactical grid before returning the outcome to the strategic map.

**trigger conditions:**
- field battle: army moves onto tile occupied by enemy army.
- sortie: defender orders a sortie while city is under siege.
- ambush (planned): high-strategy general may intercept a marching army.

**grid dimensions:**
- **15 × 9 tiles** (width × height).
- attacker deploys from the right 3 columns (x 12–14); defender deploys from the left 3 columns (x 0–2).
- centre 9 columns (x 3–11) are contested ground.
- grid tiles are generated from the strategic map tile at the battle location (terrain carries over).

```
 defender zone  |  contested ground  |  attacker zone
  [ x 0–2 ]         [ x 3–11 ]           [ x 12–14 ]
```

**unit types:**

each army deploys its troops split into up to 4 units based on officer tags:

| unit type | officer tag | move | attack range | strength |
| :--- | :--- | :---: | :---: | :--- |
| infantry | military | 2 | 1 (melee) | balanced; holds ground |
| cavalry | cavalry | 4 | 1 (charge) | high charge damage; weak vs spear |
| archers | scholar/strategist | 2 | 3 (ranged) | high range; low defence |
| naval | naval | 3 (water) | 2 | strong on river/coast only |
| siege engine | engineer | 1 | 4 (ranged) | high siege damage; immobile once placed |

if the army has no matching officer tag, all troops default to infantry.

**turn order:**
1. initiative determined by general's strategy (higher goes first). ties broken by valour.
2. each unit gets 1 action per round: move, attack, or hold.
3. player selects unit → valid move tiles highlight → player selects destination or target.
4. after all player units act, enemy AI resolves its units.
5. round ends; morale and casualty totals updated; next round begins.

**action points per unit per round:**

| action | cost | notes |
| :--- | :--- | :--- |
| move | 1 AP | move up to unit's move range |
| attack | 1 AP | attack adjacent or in-range tile |
| move + attack | 2 AP | available only to cavalry (charge) |
| hold / rally | 0 AP | skip action; +5 morale to this unit |

**battle terrain tiles:**

terrain at the battle location is drawn from the strategic map tile and expanded into a tactical pattern:

| terrain | tactical tile effect |
| :--- | :--- |
| plain | no modifier; cavalry move +1 |
| forest | −1 move to all units; archers −1 range; +10% defence for infantry in forest |
| river | impassable to non-naval units; naval +25% |
| mountain / hill | +20% defence for units on high ground; −1 move uphill |
| coast | naval free movement; others −1 move in shallow water tiles |
| pass | bottleneck tile (1-wide column); defender holds with small force |
| fortification | city wall tile (siege only); DEF reduces damage; engineer demolishes |

**battle resolution per round:**

```
damage = (attacker.troops × valour_mod × terrain_mod × element_mod × morale_mod) ÷ 100
defender.troops -= damage
defender.morale -= damage ÷ 50
```

modifiers:
- `valour_mod` = attacking officer's valour ÷ 50 (range 0.02–2.0)
- `terrain_mod` = from tactical terrain table above
- `element_mod` = essence drift multiplier of commanding officer
- `morale_mod` = attacker morale ÷ 100

**duel mechanics:**
- triggered when two generals are on adjacent tiles and a charge action is taken.
- also triggered probabilistically at battle start if valour difference > 20.
- resolution: both generals roll `valour + 1d20`; higher roll wins.
- outcomes:

| result | effect |
| :--- | :--- |
| decisive win (diff > 15) | loser dies or flees; −30 morale to loser's army |
| win (diff 5–15) | loser retreats 2 tiles; −15 morale to loser's army |
| draw (diff < 5) | both injured (−10 health); no morale change |

- vanguard tag: +10 to duel roll.
- cavalry tag: +5 to duel roll on plain terrain.
- loyalist tag: officer refuses to flee on decisive loss (stands and fights to 0 health).

**morale and rout:**
- each unit has morale (1–100); army morale = average of all unit morales.
- morale drops on: casualties, failed charges, general killed/fled, being flanked.
- if a unit's morale hits 0: unit routs — moves 3 tiles away from enemy and loses 10% troops per turn.
- if army morale hits 0: full rout — all units flee; pursuer deals +50% bonus casualties.

**battle victory conditions:**

| condition | trigger | result |
| :--- | :--- | :--- |
| rout | enemy army morale = 0 | winner holds tile; loser retreats to nearest friendly city |
| annihilation | enemy troops < 10% starting | as above; bonus ledger_log entry |
| general killed | commanding general dies in duel | −40 morale; forced rout check |
| retreat ordered | player manually retreats | army moves 2 tiles back; no penalty beyond lost ground |
| siege fall | DEF = 0 during siege | city captured; garrison eliminated |

**siege battle specifics:**
- battle grid uses fortification tiles on the defender's left edge (city wall).
- attacker must breach wall tiles (reduced by siege engine fire) before reaching city interior.
- defender can place garrison troops on wall tiles at +20% defence bonus.
- sortie: defender can send units into contested ground; if attacker morale breaks, siege lifts.

**battle outcomes recorded in ledger:**
- BATTLE_WON / BATTLE_LOST / BATTLE_DRAWN
- GENERAL_KILLED / GENERAL_FLED / DUEL_RESULT
- SIEGE_LIFTED / CITY_CAPTURED
- all troop losses, officer health changes, and morale deltas

**Godot scene structure:**

| node | purpose |
| :--- | :--- |
| BattleScene (root) | manages round loop, initiative queue, outcome dispatch |
| BattleTileMap | 15×9 tactical grid; terrain tiles from strategic map context |
| UnitLayer | unit tokens for each deployed unit (attacker + defender) |
| UILayer | action panel (move/attack/hold), morale bars, round counter |
| DuelScreen (modal) | animated duel sequence overlaid on battle scene |
| BattleLog | scrolling round-by-round event feed |

**status:** battle system not yet implemented. duel roll and morale model defined above are the initial implementation target.

---

### 6. diplomacy

inter-faction relations managed through a dedicated diplomacy phase resolved between phase A and phase B.

**relation score:**
- each pair of factions has a relation score (−100 hostile to +100 allied).
- changes based on: wars, tribute, gifts, marriages, broken agreements, shared enemies.

**diplomatic actions (cost CP or gold):**

| action | cost | effect |
| :--- | :--- | :--- |
| send tribute | gold | +relation; buys temporary non-aggression |
| propose alliance | CP + gold | formal alliance; both factions share enemy info |
| gift officer | — | send a low-loyalty officer as goodwill; risky |
| demand surrender | — | offer target faction vassalage; accepted if relation very high |
| threaten | — | −relation with target; may deter attack |
| spy | CP | place a spy in target city; reveals troop/pillar data |
| bribe officer | gold | attempt to reduce target officer's loyalty; resisted by integrity |
| marriage | — | formal alliance via officer/family link; long-term +relation |

**alliance tiers:**

| tier | threshold | benefit |
| :--- | :--- | :--- |
| neutral | 0–30 | no hostility pact; no shared intel |
| friendly | 31–60 | trade access; minor intel sharing |
| allied | 61–85 | coordinated campaigns; shared supply |
| vassal / suzerain | 86–100 | subordinate faction; tribute flows to suzerain |

**espionage:**
- spy placed in a city remains until discovered (integrity check each turn).
- active spy: reveals city pillars, garrison, officer assignments, and army positions.
- spy discovery: triggers ESPIONAGE_DETECTED event; −relation with source faction.

**status:** not yet implemented.

---

### 7. events system

phase A generates world events each turn. some are automatic (seasonal); others are random or historically scripted.

**event categories:**

| category | trigger | example |
| :--- | :--- | :--- |
| seasonal | every turn | grain signal, gold signal, order drift |
| random | probability roll | locust swarm (−AG), flood (−COM), epidemic (−population) |
| NPC actor | officer tag + probability | physician (Hua Tuo) restores officer health; sage offers advice; merchant opens trade route |
| historical | year + conditions | Yellow Turban Rebellion, Dong Zhuo coup, Battle of Guandu trigger windows |
| officer lifecycle | age / loyalty / health | officer falls ill, retires, defects, dies, has child |
| faction | relation score | rival declares war, proposes alliance, faction collapses |

**event resolution:**
- events present a choice (accept/decline) or apply automatically.
- effects are recorded in `ledger_log` with effects_json.
- some events are permanent (officer death); others are temporary (flood).

**historical scripted events (planned):**

| event | trigger window | effect |
| :--- | :--- | :--- |
| Yellow Turban Rebellion | AD 184 (epoch) | initial instability; ORD −20 in central cities |
| Dong Zhuo seizes capital | AD 189 | Luoyang/Chang'an factions shift; coalition forms |
| Battle of Guandu | AD 200 (if Yuan Shao vs Cao Cao active) | scripted army clash event |
| Red Cliffs | AD 208 (if Cao Cao marches south) | naval penalty for northern faction |

**NPC actors:**
- drawn from the officer archive by tag.
- appear as phase A choice events: "Hua Tuo arrives in [city]. Treat your wounded general? (costs 50 gold)"
- accepting applies the effect and logs the NPC visit; declining passes.

**status:** seasonal events (SEASON_SHIFT, GRAIN_SIGNAL, etc.) implemented. random, NPC, historical, and lifecycle events not yet implemented.

---

### 8. turn structure (full)

each turn = one month. phases execute in strict order; nothing is committed until phase C completes.

```
┌─────────────────────────────────────────────────────────┐
│  phase A — world update (automatic, no player input)    │
│    1. bazi clock advances                               │
│    2. seasonal deltas applied to all cities             │
│    3. essence drift recalculated for all officers       │
│    4. random / historical events generated              │
│    5. NPC actor events offered                          │
│    6. loyalty drift applied to all officers             │
│    7. army supply consumed; attrition if supply = 0     │
│    8. population growth/decline computed                │
├─────────────────────────────────────────────────────────┤
│  phase D — diplomacy (optional, player-initiated)       │
│    - send tribute, propose alliance, bribe, spy         │
│    - no CP cost; costs gold or dedicated diplomat CP    │
├─────────────────────────────────────────────────────────┤
│  phase B — player commands (CP-limited)                 │
│    - city development: ag / com / tech / ord / def      │
│    - officer: search / recruit / assign / dismiss       │
│    - military: recruit / march / siege / encamp         │
│    - commands queued; engine validates CP budget        │
├─────────────────────────────────────────────────────────┤
│  phase C — settlement (atomic SQLite transaction)       │
│    - apply all phase A / D / B deltas                   │
│    - resolve battles triggered by army movement         │
│    - resolve siege outcomes                             │
│    - advance game_clock                                 │
│    - append all events to ledger_log                    │
│    - rollback on failure; ledger unchanged              │
└─────────────────────────────────────────────────────────┘
```

**command points (CP):**
- generated by the active lord: `strategy (post-drift) ÷ 10`, min 5.
- advisor officers attached to lord provide +1 CP each (max 3 advisors).
- diplomacy actions use a separate diplomat CP pool (planned).

**seasonal deltas per city:**

| season | element | AG | COM | TECH | ORD |
| :--- | :--- | :---: | :---: | :---: | :---: |
| spring | Wood | +2 | 0 | +1 | 0 |
| summer | Fire | +1 | +2 | +2 | -1 |
| transition | Earth | 0 | 0 | 0 | +2 |
| autumn | Metal | +3 | +1 | 0 | +1 |
| winter | Water | -2 | -1 | 0 | 0 |

transition months: 3, 6, 9, 12.

---

### 9. essence drift

the clock module computes a float multiplier (0.80–1.25) applied to all officer stats each turn, based on the Wu Xing relationship between the officer's essence and the current dominant element.

| relationship | multiplier | description |
| :--- | :---: | :--- |
| peak (same element) | 1.25 | officer element matches dominant element exactly |
| nourished (dominant generates officer) | 1.15 | dominant element feeds the officer's element |
| feeding (officer generates dominant) | 1.10 | officer's element feeds the dominant |
| resistant (officer controls dominant) | 0.90 | officer's element controls the dominant |
| suppressed (dominant controls officer) | 0.80 | dominant element controls the officer's element |

Wu Xing cycles:

```
generating: Wood -> Fire -> Earth -> Metal -> Water -> Wood
controlling: Wood -x Earth -x Water -x Fire -x Metal -x Wood
```

| essence | peak season | suppressed season |
| :--- | :--- | :--- |
| Wood | spring | autumn |
| Fire | summer | winter |
| Earth | transitions | — |
| Metal | autumn | spring |
| Water | winter | summer |

effective stat = clamp(base × drift, 1, 100).

---

### 10. bazi calendar (天干地支)

tracks time using the traditional 60-unit stem-branch cycle. epoch: AD 184 (Yellow Turban Rebellion).

- 10 heavenly stems (天干): drive the year's elemental character.
- 12 earthly branches (地支): drive the month's dominant element, season, and zodiac animal.
- combined: 10 × 12 = 60-month repeating cycle.

derived each turn:
- `dominant_element`: Wu Xing element governing this month.
- `season`: spring / summer / transition / autumn / winter.
- `is_transition_month`: true for months 3, 6, 9, 12.

---

### 11. victory and legacy

no single win screen. victory is assessed from the ledger at three levels.

**victory conditions:**

| dimension | measure | assessed from |
| :--- | :--- | :--- |
| sovereignty | cities controlled at era end, weighted by strategic value | `cities.faction` history in ledger_log |
| institutional strength | officer loyalty avg × governance avg × ORD avg across all cities | officer and city records in ledger_log |
| historical judgment | ratio of just acts to brutal acts; elemental balance of policy decisions | event tags in ledger_log (JUST / BRUTAL / BALANCED) |

**era end trigger:**
- unification: one faction holds all cities → sovereignty victory.
- year 280 (era end): game scores all three dimensions; highest composite wins.
- annihilation: faction loses all cities → eliminated.

**legacy score:**
- a player who unifies through betrayal and scorched earth holds the land but scores low on historical judgment.
- a player who builds deep institutions and balanced policy may score higher than the military victor.
- ledger_log is the permanent record; all decisions are traceable.

---

## persistence

three storage layers, each matched to its concern:

| layer | format | file | concern |
| :--- | :--- | :--- | :--- |
| canonical seed data | YAML | `data/officers.yaml`, `data/cities.yaml` | human-readable archive; read once at first run |
| runtime game state | SQLite | `ledger.db` | relational ledger; atomic transactions; append-only log |
| player preferences | JSON | `user://prefs.json` | Godot user data dir; no game logic dependency |

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
| `ledger_log` | append-only event log (year, month, phase, event_type, description, effects_json) |

all pillar columns have CHECK constraints (1–100). `ledger_log` is a chronicle — victory scoring queries it, but it is not the source of truth for current game state.

### JSON (`user://prefs.json`)

written and read by Godot via `FileAccess`. contains no game state — safe to delete without affecting a saved game.

```json
{
  "display": {
    "fullscreen": false,
    "resolution": "1280x720"
  },
  "audio": {
    "master_volume": 1.0,
    "music_volume": 0.8,
    "sfx_volume": 1.0
  },
  "gameplay": {
    "last_scenario": "189ad",
    "text_speed": "normal"
  }
}
```

---

## frontend

**Godot 4** is the frontend engine. the headless modules (engine, battle, diplomacy, events, clock, ledger, archive, models) are the authoritative core; Godot renders state and dispatches player commands. YAML archives and SQLite ledger are engine-level concerns untouched by the frontend.

rationale:
- built-in scene tree, tilemap, and animation systems suit the grid-based strategy map.
- open-source, no royalties, strong 2D support for the pixel art / Han character aesthetic.
- multi-platform export (desktop, web) without licensing constraints.
- GDScript and C# are both viable for the frontend layer.

### scenes

| scene | maps to system | status |
| :--- | :--- | :--- |
| SplashScreen | — | hello world scaffold done |
| StrategicMapScreen | strategic map — tilemap, armies, faction overlays | not started |
| CityScreen | city development — pillar bars, officer slot, food/gold | not started |
| OfficerScreen | officer management — roster, stats, assignment | not started |
| ArmyScreen | army system — troop count, morale, supply, stance | not started |
| BattleScreen | tactical battle — grid, units, commands, duel animation | not started |
| DiplomacyScreen | diplomacy — faction list, relation scores, action panel | not started |
| LedgerScreen | ledger log — scrollable history, event filter | not started |
| VictoryScreen | legacy scoring — three-dimension breakdown | not started |

---

## phases

> `done` · `in progress` · `not started`

| phase | focus | status |
| :--- | :--- | :--- |
| 0 | foundation | done |
| 1 | data + turn engine | not started |
| 2 | cities | not started |
| 3 | officers | not started |
| 4 | army system | not started |
| 5 | strategic map | not started |
| 6 | tactical battle | not started |
| 7 | diplomacy + events | not started |
| 8 | victory + polish | not started |

---

### phase 0 — foundation `done`

| item | status |
| :--- | :--- |
| Godot 4 project (1280×720, pixel settings, Brewfile, Makefile) | done |
| splash screen (fade, blink prompt, 5s auto-advance, key dismiss) | done |
| main scene placeholder | done |
| design: officer + city archives (YAML, 498 officers, 30 cities) | done |
| design: bazi calendar, turn structure, SQLite schema | done |

---

### phase 1 — data + turn engine `not started`

SQLite init, YAML seeding, bazi clock, and the three-phase turn loop in GDScript.

| item | status |
| :--- | :--- |
| SQLite init (schema, WAL, seed from YAML) | not started |
| bazi calendar + essence drift | not started |
| phase A — seasonal deltas | not started |
| phase B — CP commands | not started |
| phase C — atomic settlement | not started |

---

### phase 2 — cities `not started`

city resources and economic simulation.

| item | status |
| :--- | :--- |
| city food + gold output per turn | not started |
| population growth / decline | not started |
| unrest trigger (ORD < 30) | not started |
| city screen (pillar bars, resources, governor slot) | not started |

---

### phase 3 — officers `not started`

officer lifecycle, allegiance, and assignment.

| item | status |
| :--- | :--- |
| officer allegiance table + loyalty drift | not started |
| officer assignment (governor / general / advisor) | not started |
| officer search + recruitment | not started |
| officer experience + stat growth | not started |
| officer age / health / death | not started |

---

### phase 4 — army system `not started`

raising, moving, and supplying armies.

| item | status |
| :--- | :--- |
| army model (troops, morale, supply, stance) | not started |
| troop recruitment from cities | not started |
| army movement + supply line + attrition | not started |
| siege (DEF degradation, city fall) | not started |

---

### phase 5 — strategic map `not started`

China campaign map scene.

| item | status |
| :--- | :--- |
| Godot TileMap (terrain tiles, faction colour overlay) | not started |
| city nodes (size by population tier) | not started |
| army tokens + movement on map | not started |
| city count expansion (30 → 41+) | not started |

---

### phase 6 — tactical battle `not started`

battle scene and turn-based resolution.

| item | status |
| :--- | :--- |
| battle grid (15×9, deployment zones) | not started |
| unit types + movement + damage formula | not started |
| morale (break and rout) | not started |
| officer duels | not started |
| Godot battle scene | not started |

---

### phase 7 — diplomacy + events `not started`

inter-faction relations and world events.

| item | status |
| :--- | :--- |
| faction relation scores + diplomacy actions | not started |
| random events (flood, locust, epidemic) | not started |
| NPC event actors | not started |
| historical scripted events | not started |

---

### phase 8 — victory + polish `not started`

scoring, remaining scenes, multi-era.

| item | status |
| :--- | :--- |
| victory scoring (territory, institutional, legacy) | not started |
| era-end trigger (unification or year 280) | not started |
| remaining Godot scenes (officer, diplomacy, ledger) | not started |
| multi-era framework | not started |
