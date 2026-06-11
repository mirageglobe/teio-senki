# design: 帝王战纪：三国录

> "Sovereignty through the Ledger, Strategy through the Elements."

## overview

**帝王战纪：三国录** is a turn-based grand strategy simulation set during China's Three Kingdoms era (AD 184–280), beginning at the Yellow Turban Rebellion epoch. the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft, elemental alignment, and long-term institutional building.

victory is measured by territory — one faction controls all cities. institutional and legacy scoring are deferred to post-release.

---

## platform targets

| platform                          | priority  | input            | notes                                                                   |
| :-------------------------------- | :-------: | :--------------- | :---------------------------------------------------------------------- |
| desktop (Windows / macOS / Linux) | primary   | keyboard + mouse | full feature set; reference platform for development                    |
| mobile (Android / iOS)            | secondary | touch            | same codebase; UI must be touch-friendly at 1080p portrait or landscape |

mobile is a post-milestone-8 port target. do not design around mobile constraints during active development — but do not make decisions that actively block a future port (no fixed pixel coordinates, no keyboard-only flows without a touch fallback).

## visual constraints

- **2D only** — no 3D geometry. Ebitengine is 2D-native; depth illusion via sprite shading is acceptable but not required.
- **pixel art** — all sprites, tiles, and UI elements are pixel art at a constrained palette. no high-resolution or vector assets. target tile size: 16×16 or 32×32 px scaled up.
- **low asset cost** — every visual element must be cheap to produce: flat sprites, palette-swaps for factions, minimal animation (idle shimmer, flag wave). no skeletal rigs, no particle systems beyond simple dust/smoke.
- **speed first** — all screens must feel instant. turn resolution, scene transitions, and UI interactions target < 100ms response. no loading screens between strategic map and city/army overlays.

## ux expectations

- **keyboard-driven (desktop)** — all core actions reachable without a mouse. mouse supported as primary pointing device.
- **touch-friendly (mobile)** — tap replaces click; long-press replaces right-click; panels must be large enough for finger targets (≥ 44px).
- **minimal UI chrome** — information density over decoration. panels appear on demand; the map is always the primary surface.
- **no tutorials** — the game communicates through design, tooltips, and the ledger log. no guided onboarding flow.

### keyboard bindings (TUI — cycle B)

canonical keybindings for the Bubble Tea frontend. all developers must use this table; do not invent bindings ad hoc.

| key       | action                                            | scope                 |
| :-------- | :------------------------------------------------ | :-------------------- |
| `e`       | end turn (advance cycle A → B → C)                | global (Playing)      |
| `m`       | open strategic map focus                          | global                |
| `c`       | open city panel for selected city                 | map                   |
| `o`       | open officer roster panel                         | map                   |
| `a`       | open army panel for selected army                 | map                   |
| `d`       | open diplomacy panel                              | map                   |
| `l`       | open ledger log                                   | global                |
| `↑ ↓ ← →` | move map cursor / scroll panel                    | map / panels          |
| `enter`   | confirm selection / issue command                 | panels                |
| `esc`     | close panel / cancel command; dismiss command bar | panels / global       |
| `b`       | issue build command (pillar select)               | city panel (seasonal) |
| `r`       | recruit troops                                    | city panel (seasonal) |
| `s`       | search for officers                               | city panel            |
| `g`       | assign governor                                   | officer panel         |
| `h`       | assign general                                    | officer panel         |
| `v`       | assign advisor                                    | officer panel         |
| `t`       | send tribute                                      | diplomacy panel       |
| `p`       | propose alliance                                  | diplomacy panel       |
| `w`       | threaten faction                                  | diplomacy panel       |
| `q`       | quit to main menu                                 | global                |
| `:`       | open command bar                                  | global (Playing)      |

mouse: clicking a city node or army token selects it and opens the relevant panel. all keyboard actions have a mouse/touch equivalent for mobile.

### command bar

press `:` in any Playing state to open the command bar at the bottom of the TUI. type a `/command` and press `enter` to issue it directly, bypassing panel navigation. `esc` dismisses the bar without issuing a command.

the command bar is a power-user shortcut — it does not replace panel navigation. both paths call identical engine methods and share the same CP validation logic.

**command reference:**

| command      | syntax                                                 | cp  | availability | effect                                                       |
| :----------- | :----------------------------------------------------- | :-: | :----------- | :----------------------------------------------------------- |
| `/develop`   | `/develop <city_id> <pillar>`                          | 1   | seasonal     | grow a pillar (AG / COM / DEF) by one tick                   |
| `/recruit`   | `/recruit <city_id> <amount>`                          | 1   | seasonal     | raise troops from city population (gold cost applies)        |
| `/train`     | `/train <city_id>`                                     | 1   | monthly      | increase garrison morale                                     |
| `/search`    | `/search <city_id>`                                    | 1   | monthly      | chance to discover an unrecruited officer in the region      |
| `/assign`    | `/assign <officer_id> <role> [target_id]`              | 1   | monthly      | assign officer as governor, general, advisor, or successor   |
| `/reward`    | `/reward <officer_id>`                                 | 0   | monthly      | spend gold to boost officer loyalty                          |
| `/transport` | `/transport <from_city> <to_city> <resource> <amount>` | 1   | monthly      | move food or gold between cities                             |
| `/march`     | `/march <army_id> <x> <y>`                             | 1   | monthly      | queue march order; applied in cycle C settlement             |
| `/encamp`    | `/encamp <army_id>`                                    | 1   | monthly      | fortify position; −50% incoming damage; no advance           |
| `/tribute`   | `/tribute <faction> <gold>`                            | 0   | monthly      | send gold to improve relation score                          |
| `/alliance`  | `/alliance <faction>`                                  | 0   | monthly      | propose alliance (costs 200 gold; accepted if relation ≥ 20) |
| `/threaten`  | `/threaten <faction>`                                  | 0   | monthly      | threaten faction; may trigger war if relation < −30          |

> commands marked **seasonal** are only available on the first turn of each season (months 1, 4, 7, 10).

---

## domain models

### Officer

| field      | type       | notes                                                                    |
| :--------- | :--------- | :----------------------------------------------------------------------- |
| id         | string     | unique identifier                                                        |
| name       | string     | historical name                                                          |
| title      | string     | honorific / rank                                                         |
| essence    | Element    | elemental root; drives seasonal drift multiplier                         |
| strategy   | int 1–100  | army throughput, CP generation                                           |
| valour     | int 1–100  | personal duels, vanguard charges                                         |
| governance | int 1–100  | city yield, trade efficiency                                             |
| integrity  | int 1–100  | loyalty resilience, bribery immunity                                     |
| loyalty    | int 1–100  | bond to current lord; drifts each turn based on salary, battles, essence |
| faction    | string     | currently serving faction; update on defection                           |
| health     | int 1–100  | officer health; declines with age, wounds, illness                       |
| age        | int        | current age; officers die of old age or in battle                        |
| experience | int 0–9999 | accumulated XP; thresholds unlock stat growth                            |
| city_id    | string\    | null                                                                     |
| army_id    | string\    | null                                                                     |
| tags       | list[Tag]  | specialist classifiers                                                   |

### City

| field       | type      | notes                                                             |
| :---------- | :-------- | :---------------------------------------------------------------- |
| id          | string    | unique identifier                                                 |
| name        | string    | romanised name                                                    |
| chinese     | string    | Han character name                                                |
| region      | string    | strategic region                                                  |
| terrain     | Terrain   | terrain type                                                      |
| x, y        | int       | grid position (x: 0–19, y: 0–14)                                  |
| population  | int       | static for MVP; caps recruitment per turn                         |
| faction     | string    | controlling faction                                               |
| governor_id | string\   | null                                                              |
| garrison    | int       | troops stationed for city defence                                 |
| agriculture | int 1–100 | AG pillar — active                                                |
| commerce    | int 1–100 | COM pillar — active                                               |
| technology  | int 1–100 | TECH pillar — active in yield formula; BUILD command deferred     |
| order       | int 1–100 | ORD pillar — active in corruption formula; BUILD command deferred |
| defense     | int 1–100 | DEF pillar — active                                               |
| food        | int       | food stockpile in units; army upkeep drawn here                   |
| gold        | int       | treasury; recruitment and diplomacy costs drawn here              |

### Army

| field       | type         | notes                                                         |
| :---------- | :----------- | :------------------------------------------------------------ |
| id          | string       | unique identifier                                             |
| faction     | string       | owning faction                                                |
| general_id  | string       | commanding officer (strategy drives formation size, movement) |
| officer_ids | list[string] | attached subordinate officers                                 |
| troops      | int          | current troop count                                           |
| morale      | int 1–100    | battle effectiveness multiplier; drops on retreat/defeat      |
| supply      | int          | turns of food remaining; 0 = attrition                        |
| x, y        | int          | current map position                                          |
| stance      | Stance       | idle / march / siege / encamp                                 |

### Element

Wood, Fire, Earth, Metal, Water

### Terrain

Plain, Mountain, Forest, River, Coast, Pass

### Tag

**classification tags**:

| tag      | meaning                                           |
| :------- | :------------------------------------------------ |
| civil    | civil administrator                               |
| military | military commander                                |
| lord     | eligible sovereign; generates CP; +20% governance |

**specialist tags**:

| tag        | bonus           | situation                    |
| :--------- | :-------------- | :--------------------------- |
| naval      | +25% strategy   | river / sea combat           |
| cavalry    | +20% valour     | open-terrain charge          |
| vanguard   | +15% valour     | duel or charge               |
| scholar    | +20% governance | academy / civil admin        |
| diplomat   | +20% integrity  | alliance / espionage defence |
| engineer   | +20% governance | siege / infrastructure       |
| merchant   | +25% governance | active trade route           |
| strategist | +15% strategy   | tactical positioning         |
| loyalist   | +25% integrity  | enemy bribery attempt        |
| pioneer    | +15% governance | newly annexed territory      |

---

## data archives

| archive              | count        | notes                                                      |
| :------------------- | :----------- | :--------------------------------------------------------- |
| `data/officers.yaml` | 498 officers | Wei, Shu, Wu, and independent factions                     |
| `data/cities.yaml`   | 30 cities    | fixed grid positions, all five pillars, faction assignment |

target city count is 41–47. current 30 covers the core theatre; southern (Jiao Zhou) and far-north (Liaodong) cities are the main gaps.

archives are the canonical human-readable source. they seed the in-memory ledger on new game start. subsequent launches load from `save.json` instead.

---

## scenario & player setup

before the game loop begins, the player must define the starting conditions of the simulation.

### 1. scenario selection

scenarios define the historical epoch, including the distribution of cities, the status of factions, and the available pool of officers.

| scenario                | epoch  | description                                              |
| :---------------------- | :----- | :------------------------------------------------------- |
| Yellow Turban Rebellion | AD 184 | the fall of the Han; Zhang Jiao's uprising               |
| Dong Zhuo's Rise        | AD 189 | chaos in the capital; the anti-Dong Zhuo coalition forms |
| Rivalry of Lords        | AD 194 | the death of Dong Zhuo; independent lords vie for power  |
| Battle of Guandu        | AD 200 | the clash between Cao Cao and Yuan Shao                  |
| Three Kingdoms          | AD 220 | the formal division of the empire into Wei, Shu, and Wu  |

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

## game mechanics

### core loop

each session is a sequence of turns. one turn = one calendar month. the player's goal is to unify all cities under their faction before rival sovereigns do.

two cadences run simultaneously:

| cadence      | frequency                                         | governs                                                                       |
| :----------- | :------------------------------------------------ | :---------------------------------------------------------------------------- |
| **monthly**  | every turn                                        | army movement, battle resolution, supply checks, loyalty drift, essence drift |
| **seasonal** | every 3 turns (spring / summer / autumn / winter) | city yield (food & gold), pillar development growth, troop recruitment        |

this keeps tactical decisions (where to march, when to fight) active every month while preventing administrative exhaustion from managing city economics on the same cadence.

```
each turn (1 month):
  A — world update  →  clock advances; essence drifts; loyalty decays; supply / attrition checks
                        [seasonal only] city yields calculated; pillar growth applied
  B — player commands  →  spend CP on movement / combat / officer / diplomacy orders every turn
                           [seasonal only] recruit and city develop commands unlocked
  C — settlement  →  apply all deltas sequentially; resolve battles; advance clock
                      [seasonal only] apply yield and recruitment results; check season boundary
```

### command points (CP)

CP is the player's action budget per turn. it limits how much can be done in a single month.

| source                           | CP granted                                 |
| :------------------------------- | :----------------------------------------- |
| sovereign's `strategy` stat      | base CP = `strategy ÷ 10` (min 1)          |
| advisor officer attached to lord | +1–3 CP per advisor (scales with strategy) |

CP is consumed by commands in cycle B. unspent CP is lost — it does not carry over. see the [command reference](#command-bar) for per-command CP costs.

### player actions (cycle B)

each month the player issues commands via panel navigation or the [command bar](#command-bar). all commands are validated against the CP budget before settlement. the command bar table is the canonical reference for syntax, CP cost, and availability.

### elemental alignment (wu xing)

every officer has an `essence` element (Wood / Fire / Earth / Metal / Water). the sovereign's essence sets the faction's elemental identity for the scenario.

- **seasonal drift**: each season amplifies or dampens stat effectiveness based on element cycles. spring boosts Wood officers; winter boosts Water; etc.
- **resonance bonus**: officers whose essence aligns with the sovereign's gain +1 to +3 loyalty per turn passively.
- **element cycle**: the nourishing cycle (Wood → Fire → Earth → Metal → Water → Wood) and controlling cycle (Wood controls Earth; Fire controls Metal; etc.) determine drift direction and magnitude.
- **bazi clock**: time advances through the 60-unit stem-branch cycle. the heavenly stem and earthly branch of the current month determine which elements are in ascendancy, affecting all drift calculations simultaneously.

### resource economy

two resources sustain the state, on different cadences:

| resource | generated                     | cadence                        | consumed                                 | cadence     |
| :------- | :---------------------------- | :----------------------------- | :--------------------------------------- | :---------- |
| **food** | `AG × season_delta` per city  | **seasonal** (once per season) | army upkeep (`troops ÷ 100` per army)    | **monthly** |
| **gold** | `COM × season_delta` per city | **seasonal** (once per season) | officer salaries, recruitment, diplomacy | **monthly** |

seasonal yield lands as a lump sum at the season boundary (month 3, 6, 9, 12). monthly draws (upkeep, salaries) reduce the stockpile each turn. this creates a natural rhythm: build reserves in peaceful seasons, campaign on what you have.

**officer salary formula:**

salary drawn from faction gold stockpile each cycle C:

```
salary_cost_per_officer = floor(officer.strategy + officer.valour + officer.governance) ÷ 30
```

minimum 1 gold per officer per turn regardless of stats. unassigned officers in pool still draw salary. if gold stockpile < total salary obligation: officers are ranked by loyalty descending; payment fills from top until gold is exhausted. officers with loyalty < 50 whose salary is skipped trigger the "unpaid" drift penalty that turn.

food deficit → army attrition (monthly). gold deficit → loyalty decay on unpaid officers (monthly).

### officer loyalty

officers remain loyal if treated well. loyalty drifts each turn:

- unpaid salary: −1 / turn
- salary paid: +1 / turn
- battle won: +2 / battle lost: −3
- elemental resonance with sovereign: +1 to +3
- rival bribery attempt (resisted by integrity): −5 to −20
- loyalty < 20 → defection risk next turn

### victory and defeat

**win:** one faction holds all cities → sovereignty victory. checked at end of every cycle C.

**defeat:** sovereign dies with no assigned successor, or all cities lost.

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

| region                           | 州       | rough tile zone  | key cities                     |
| :------------------------------- | :------ | :--------------- | :----------------------------- |
| Si Li (capital)                  | 司隸      | x 6–11, y 9–11   | Luoyang, Chang'an, Tongguan    |
| Ji Zhou (Hebei)                  | 冀州      | x 8–13, y 11–14  | Ye, Beiping                    |
| You Zhou (north-east)            | 幽州      | x 12–17, y 12–14 | Youzhou, Liaodong              |
| Bing Zhou (north-west)           | 并州      | x 5–9, y 11–13   | Jinyang                        |
| Liang Zhou (far west)            | 涼州      | x 1–6, y 8–12    | Tianshui, Wuwei, Dunhuang      |
| Yu Zhou / Xu Zhou (central-east) | 豫州 / 徐州 | x 10–16, y 7–10  | Xuchang, Xuzhou, Xiapi         |
| Yang Zhou (south-east)           | 揚州      | x 11–17, y 4–8   | Jianye, Hefei, Kuaiji          |
| Jing Zhou (central-south)        | 荊州      | x 7–12, y 3–8    | Xiangyang, Jiangling, Changsha |
| Yi Zhou (south-west / Shu)       | 益州      | x 2–8, y 2–8     | Chengdu, Hanzhong, Jiamenguan  |
| Jiao Zhou (far south)            | 交州      | x 6–12, y 0–3    | Nanhai, Jianning (planned)     |

**map layers:**

| layer             | content                                                                             |
| :---------------- | :---------------------------------------------------------------------------------- |
| terrain           | base tile per cell: plain, mountain, forest, river, coast, pass                     |
| geography         | fixed overlays: Yellow River, Yangtze, Qinling range, Taihang range                 |
| faction territory | coloured region fill per controlling faction; updates on city capture               |
| city nodes        | icon at city tile: size = population tier (small/medium/large); colour = faction    |
| army tokens       | unit marker at army position; number = troop count tier                             |
| season overlay    | tint shift per season: green (spring), amber (summer), gold (autumn), grey (winter) |
| ui overlay        | city name labels, army faction flags, selected unit highlight                       |

**city node tiers (population):**

| tier    | population | visual                 |
| :------ | :--------- | :--------------------- |
| village | < 100k     | small dot              |
| town    | 100k–250k  | medium icon            |
| city    | 250k–400k  | large icon             |
| capital | > 400k     | large icon + wall ring |

**navigation:**
- camera pans with arrow keys or drag; zooms with scroll wheel.
- clicking a city node → opens CityScreen overlay.
- clicking an army token → opens ArmyScreen overlay.
- two opposing armies on adjacent tiles → march order triggers battle on cycle C settlement.
- end-turn button advances all cycles and re-renders map state.

**movement:**
- `/march <army_id> <x> <y>` is issued in cycle B (player commands); movement is applied during cycle C settlement.
- movement range per turn: `floor(general.strategy ÷ 20)`, min 1, max 5 tiles.
- terrain movement costs (tiles per step):

| terrain  | land army | naval army |
| :------- | :-------: | :--------: |
| plain    | 1         | —          |
| forest   | 2         | —          |
| river    | 2         | 1          |
| coast    | 2         | 1          |
| mountain | 3         | —          |
| pass     | 2         | —          |

- naval armies (general has naval tag) move freely on river/coast; 3× cost on land tiles.
- army entering a tile with an enemy army triggers battle (resolved in cycle C settlement).
- army adjacent to enemy city with siege stance: DEF degrades each turn.

**supply line:**
- supply check = Chebyshev distance to nearest friendly city ≤ 5 tiles.
- if out of range: supply countdown begins (5 turns before attrition). no pathfinding required.

**data requirements:**
- `armies` table: faction, general_id, troops, morale, supply, x, y, stance.
- `map_tiles` table: x, y, terrain (for non-city tiles beyond city list).

**status:** city grid and terrain types exist in archive. army model, map_tiles table, TUI map rendering, and movement not yet implemented.

---

### 2. city development

each city is an economic unit that generates food, gold, and manpower each turn. the player allocates commands to grow pillars and sustain their state.

**core city actions:**

| action                   | cost        | effect                        | strategic value                 |
| :----------------------- | :---------- | :---------------------------- | :------------------------------ |
| **Develop (AG/COM/DEF)** | 1 CP        | pillar growth                 | core long-term growth           |
| **Recruit**              | 1 CP + gold | increases garrison            | essential for expansion/defense |
| **Train**                | 1 CP        | increases garrison morale     | vital for battle readiness      |
| **Search**               | 1 CP        | discover unrecruited officers | adds personality & talent pool  |
| **Reward**               | gold        | +loyalty to governor/assigned | prevents defection              |
| **Transport**            | 1 CP        | moves food/gold to other city | allows logistical depth         |

**derived outputs:**
- **seasonal** (month 3 / 6 / 9 / 12): `food += AG × season_delta`; `gold += COM × season_delta`
- **monthly** (every turn): `food -= army_upkeep`; `gold -= officer_salaries`
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

| role       | effect                                                             |
| :--------- | :----------------------------------------------------------------- |
| governor   | applies governance bonus to city pillars; manages food/gold output |
| general    | commands an army; strategy drives CP allowance for battle commands |
| advisor    | attached to lord; provides passive bonus to CP budget              |
| unassigned | in the officer pool; earns salary but provides no bonus            |

**experience (XP):**
- officers gain XP through active service; unassigned officers gain nothing.
- XP sources:

| action                                 | XP gained             |
| :------------------------------------- | :-------------------- |
| battle won (as general)                | +200                  |
| battle won (as subordinate)            | +80                   |
| duel won                               | +150                  |
| governing a city for 1 year (12 turns) | +50                   |
| successful search / diplomacy action   | +30                   |
| city developed under governance        | +10 per pillar raised |

- XP thresholds unlock +1 to a relevant stat (strategy for generals, governance for governors) and may trigger a rank promotion event in the ledger.

| threshold | tier    | reward                               |
| :-------- | :------ | :----------------------------------- |
| 500       | novice  | +1 to primary stat                   |
| 1500      | veteran | +2 to primary stat                   |
| 3500      | master  | +3 to primary stat; new tag eligible |

**search mechanic:**
- `/search <city_id>` command: costs 1 CP; chance to discover an unrecruited officer in the region.
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
- `/recruit <city_id> <amount>` command: costs gold and 1 CP; draws from city population.
- max recruitable per turn: `population × 0.05`.
- trained troops cost more but have higher base morale.

**supply and logistics:**
- each army consumes food per turn: `troops ÷ 100` units from the nearest friendly city.
- if supply drops to 0: morale −5 per turn; troops −2% per turn (attrition).
- supply line: Chebyshev distance to nearest friendly city ≤ 5 tiles; out of range = countdown begins.

**army stances:**

| stance | effect                                               |
| :----- | :--------------------------------------------------- |
| idle   | stationed; no movement; full resupply                |
| march  | moving toward target tile                            |
| siege  | attacking a city; DEF degrades each turn             |
| encamp | fortified position; −50% incoming damage; no advance |

**movement:**
- armies move via player orders issued in cycle B: `/march <army_id> <x> <y>`; applied in cycle C settlement.
- movement range per turn: `general.strategy ÷ 20`, min 1, max 5 tiles.
- entering a tile occupied by an enemy army triggers battle resolution in cycle C.

**status:** army model not yet implemented. troop count tracked only as garrison on cities.

---

### 5. tactical battle (auto-resolve)

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

diplomatic actions are issued in cycle B alongside other player commands. gold cost only — no CP consumed.

**relation score:**
- each ordered pair of factions holds a score in [−100, +100]. scores are not symmetric: A→B and B→A are tracked separately and drift independently.
- score floor/ceiling: clamped at −100 / +100 after every cycle C.

**diplomatic actions:**

| action           | gold cost | relation effect                    | notes                                                          |
| :--------------- | :-------: | :--------------------------------- | :------------------------------------------------------------- |
| send tribute     | 100–500   | +10 to +30 (scales with amount)    | one-way; increases target's score toward player only           |
| propose alliance | 200       | +20 if accepted; −5 if rejected    | target AI accepts if relation ≥ 20; rejects if at war          |
| threaten         | 0         | −20 to target; +10 self-confidence | may trigger immediate war declaration if target relation < −30 |

**relation thresholds:**

| score      | state    | effect                                                            |
| :--------- | :------- | :---------------------------------------------------------------- |
| ≥ 60       | allied   | non-aggression; shared border armies don't trigger battle         |
| 20–59      | friendly | tribute costs reduced 20%                                         |
| −19 to 19  | neutral  | no effect                                                         |
| −20 to −59 | hostile  | march orders toward their cities +1 CP cost                       |
| ≤ −60      | war      | armies auto-battle on tile contact; no diplomatic action possible |

**passive drift per turn:**
- no active contact: −1 (relations decay toward neutral without maintenance).
- shared border (armies adjacent): −2.
- existing alliance: +1.

**status:** not yet implemented.

---

### 7. events system

cycle A generates world events each turn. initially limited to seasonal shifts and resource signals.

**status:** seasonal events implemented. random and historical events deferred.

---

### 8. turn structure (full)

each turn = one month. three cycles execute in strict order.

```
┌─────────────────────────────────────────────────────────────────┐
│  cycle A — world update (automatic, no player input)            │
│    [monthly]  1. bazi clock advances                            │
│    [monthly]  2. essence drift recalculated for all officers    │
│    [monthly]  3. loyalty drift calculated                       │
│    [monthly]  4. army supply / attrition computed               │
│    [seasonal] 5. city yields calculated (AG → food, COM → gold) │
│    [seasonal] 6. pillar growth deltas staged                    │
│    [monthly]  7. random / historical events generated           │
├─────────────────────────────────────────────────────────────────┤
│  cycle B — player commands (CP-limited)                         │
│    [monthly]  march / siege / encamp / diplomacy / officer ops  │
│    [seasonal] develop city pillars (AG / COM / DEF)             │
│    [seasonal] recruit troops                                    │
│    - commands queued; engine validates CP budget                │
├─────────────────────────────────────────────────────────────────┤
│  cycle C — settlement                                           │
│    [monthly]  apply movement; resolve battles (3–7 day rounds)  │
│    [monthly]  resolve siege outcomes                            │
│    [seasonal] apply yield lump sums to city stockpiles          │
│    [seasonal] apply pillar growth; apply recruitment            │
│    [monthly]  deduct upkeep (food) and salaries (gold)          │
│    [monthly]  advance game_clock; append all events to log      │
└─────────────────────────────────────────────────────────────────┘
```

---

### 9. essence drift

each officer's effective stats are modified each cycle A by a drift multiplier derived from their `essence` element and the current season's dominant element.

**drift multiplier table (officer essence vs. season element):**

| officer essence | resonant season | multiplier | controlling season                 | multiplier |
| :-------------- | :-------------- | :--------: | :--------------------------------- | :--------: |
| Wood            | spring          | 1.20       | autumn (Metal controls Wood)       | 0.85       |
| Fire            | summer          | 1.20       | winter (Water controls Fire)       | 0.85       |
| Earth           | late summer     | 1.20       | spring (Wood controls Earth)       | 0.85       |
| Metal           | autumn          | 1.20       | summer (Fire controls Metal)       | 0.85       |
| Water           | winter          | 1.20       | late summer (Earth controls Water) | 0.85       |

- neutral seasons (neither resonant nor controlling): multiplier = 1.00.
- sovereign's essence sets faction alignment: officers whose essence matches the sovereign gain an additional +0.05 multiplier (resonance bonus).
- effective stat = `clamp(base × drift_multiplier, 1, 100)` — applied to strategy, valour, governance for all calculations in that turn.

**nourishing cycle (生, shēng):** Wood → Fire → Earth → Metal → Water → Wood
**controlling cycle (克, kè):** Wood controls Earth; Earth controls Water; Water controls Fire; Fire controls Metal; Metal controls Wood.

---

### 10. bazi calendar (天干地支)

tracks time using the traditional 60-unit stem-branch cycle. epoch: AD 184 (Jiǎ Zǐ year — stem 1, branch 1).

**heavenly stems (天干, tiān gān) — 10 total:**

| index | stem   | element | polarity |
| :---: | :----- | :------ | :------- |
| 1     | 甲 Jiǎ  | Wood    | yang     |
| 2     | 乙 Yǐ   | Wood    | yin      |
| 3     | 丙 Bǐng | Fire    | yang     |
| 4     | 丁 Dīng | Fire    | yin      |
| 5     | 戊 Wù   | Earth   | yang     |
| 6     | 己 Jǐ   | Earth   | yin      |
| 7     | 庚 Gēng | Metal   | yang     |
| 8     | 辛 Xīn  | Metal   | yin      |
| 9     | 壬 Rén  | Water   | yang     |
| 10    | 癸 Guǐ  | Water   | yin      |

**earthly branches (地支, dì zhī) — 12 total, map to seasons:**

| index | branch | season      | element |
| :---: | :----- | :---------- | :------ |
| 1     | 子 Zǐ   | winter      | Water   |
| 2     | 丑 Chǒu | late winter | Earth   |
| 3     | 寅 Yín  | spring      | Wood    |
| 4     | 卯 Mǎo  | spring      | Wood    |
| 5     | 辰 Chén | late spring | Earth   |
| 6     | 巳 Sì   | summer      | Fire    |
| 7     | 午 Wǔ   | summer      | Fire    |
| 8     | 未 Wèi  | late summer | Earth   |
| 9     | 申 Shēn | autumn      | Metal   |
| 10    | 酉 Yǒu  | autumn      | Metal   |
| 11    | 戌 Xū   | late autumn | Earth   |
| 12    | 亥 Hài  | winter      | Water   |

**clock mechanics:**
- one month = one turn. the game advances one branch per turn (12 branches = 1 year).
- stems advance every 2 branches, cycling every 10 branches. the full stem-branch pair repeats every 60 turns (5 years).
- the season dominant element is read from the earthly branch of the current month and fed directly into the essence drift calculation.
- `internal/core/clock` exposes: `Advance()`, `Year() int`, `Month() int`, `Stem() int`, `Branch() int`, `SeasonElement() Element`.

---

### 11. victory

**win condition:** one faction holds all cities → sovereignty victory. checked at end of each cycle C.

**defeat condition:** sovereign dies with no assigned successor, or all cities lost.

**max turn limit (soft cap):** 360 turns (30 years). if no faction holds all cities at turn 360, the game ends with a score-based outcome: the faction controlling the most cities wins. ties broken by total population of held cities; further ties broken by gold stockpile.

**succession mechanic:**

the player may assign one eligible officer as heir at any point during cycle B:

```
/assign <officer_id> successor
```

eligibility rules:
- officer must have the **lord** tag.
- officer must currently serve the player's faction (not captured or defected).
- officer must have loyalty ≥ 50.
- only one successor can be assigned at a time; issuing the command again replaces the previous heir.

on sovereign death, the assigned successor immediately becomes the new sovereign:
- inherits the faction, all cities, and all armies.
- base CP recalculated from successor's strategy stat.
- faction elemental alignment updates to successor's essence.
- ledger logs a SUCCESSION event.

if the sovereign dies with no valid successor: defeat is triggered immediately.

**simultaneous event resolution order (cycle C):**

when multiple events of the same type occur in the same cycle C, apply in this fixed priority order:

1. supply and attrition checks (armies out of range take damage first)
2. battle resolution — armies are sorted by (attacker.troops desc, attacker.general.valour desc); the largest attacking force resolves first
3. if two armies contact each other on the same tile on the same cycle (mutual advance): the army with higher `general.valour` attacks first; if equal, resolve simultaneously (both armies take one round of damage before either retreats)
4. city capture checks (after all battles resolve)
5. victory / defeat check — if two factions hit 0 cities simultaneously in the same cycle, the faction that captured the last city wins; if both lost their last city to the same attacker in the same cycle, the attacker wins
6. yield and recruitment (seasonal only)
7. loyalty and salary settlement
8. ledger log append

---

## ai sovereign design

every faction not controlled by the player is governed by an AI sovereign. the AI runs in cycle B as a pure function — headless-testable, no UI dependency.

**faction ai record (two fields on ledger faction entry):**

| field             | type    | notes                                    |
| :---------------- | :------ | :--------------------------------------- |
| `ai_intelligence` | int 1–4 | how well the AI evaluates the game state |
| `ai_behaviour`    | string  | primary decision-making priority         |

**intelligence levels:**

| level | name      | description                                                                                          |
| :---: | :-------- | :--------------------------------------------------------------------------------------------------- |
| 1     | reckless  | random command selection; ignores supply and debt; no planning horizon                               |
| 2     | standard  | basic heuristics — build economy when safe, recruit before attacking, target weakest neighbour       |
| 3     | shrewd    | evaluates city values, maintains supply lines, coordinates army timing across turns                  |
| 4     | brilliant | optimises essence drift timing, anticipates player moves, defers attack until overwhelming advantage |

**behaviour profiles (priority weights):**

| behaviour     | military | economy | diplomacy | character                                                      |
| :------------ | :------: | :-----: | :-------: | :------------------------------------------------------------- |
| aggressive    | high     | low     | low       | attacks first; strikes before enemy consolidates               |
| defensive     | low      | high    | medium    | fortifies DEF; avoids offensive wars; counter-attacks only     |
| expansionist  | high     | medium  | low       | rapid territory grab; spreads resources thin                   |
| diplomatic    | low      | medium  | high      | prefers alliances and tribute; fights only when cornered       |
| economic      | low      | high    | medium    | maximises AG/COM before raising armies; slow but powerful late |
| opportunistic | medium   | medium  | low       | targets the weakest neighbour each turn; retreats when losing  |

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

| faction           | intelligence | behaviour     |
| :---------------- | :----------: | :------------ |
| Dong Zhuo         | 3            | aggressive    |
| Cao Cao           | 4            | opportunistic |
| Yuan Shao         | 3            | expansionist  |
| Liu Biao          | 2            | defensive     |
| Sun Jian          | 3            | aggressive    |
| Liu Zhang         | 2            | economic      |
| independent lords | 1–2          | defensive     |
