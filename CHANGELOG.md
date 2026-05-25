# Changelog

All notable changes to 帝王战纪：三国录 are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/).

---

## [0.3.0] — 2026-05-25

### added
- peripheral cities: Lelang (Korea/Han), Gungnaeseong (Goguryeo), Saro (Silla), Ito (Wa), Yamatai (Wa), Izumo (Wa), Jiuzhen (Vietnam/Han)
- `cityGeo` entries wired for all new cities — render as map dots
- `ledger_test.go`: 9 tests covering New, LoadData, Lords, GetOfficer, GetCity/SetCity, SortedCities, Log, State
- `sovereign_test.go`: 13 tests covering StartTurn, QueueCommand (accept/reject/deduct), SettleTurn (BUILD_AG/COM/DEF, queue clear, resource accumulation, stat cap)
- `internal/version` package exposing `Current` constant
- version string rendered on splash screen

### changed
- selected city highlighted in magenta on strategic map during cycle B

---

## [0.2.1] — 2026-05-24

### added
- 9 terrain types: plain, mountain, forest, river, coast, pass, marsh, hills, steppe
- terrain section added to SPEC.md with movement multipliers and art notes
- forest rendered in light sage green; hills in tan; steppe in muted yellow-green; marsh in teal
- map zoom toggle idea added to SPEC ideas section

### changed
- map bounds narrowed to 95–136°E / 18–50°N (removes empty NW quadrant)
- seasonal border colours brightened and bolded
- city pulse range constrained to white–yellow (255→226) — no overlap with terrain browns
- coast/border lines bolded for legibility

### fixed
- Gulf of Tonkin sea region no longer floods Indochina landmass
- island land (Taiwan, Japan, Korea) correctly excluded from sea fill

---

## [0.2.0] — 2026-05-23

### added
- persistent braille strategic map panel rendered alongside all game screens
- Japan (Kyushu, Shikoku, Honshu) and Taiwan landmass polygons
- pulsing city dots on map (brightness cycles ~1 s half-period)
- seasonal map border colours (Spring: green, Summer: red, Autumn: gold, Winter: cyan)
- map projected at 73–136°E / 18–55°N with ray-cast land/sea masking

---

## [0.1.1] — 2026-05-22

### added
- fixed margins (`padTop`, `padLeft`, `padBottom`) on all TUI screens
- persistent header and footer across all screens
- two-column layout: map left, content right (briefing, cycle A/B/C)
- typewriter splash animation

### changed
- map resized to 40×20 braille chars (80×80 pixel grid)

---

## [0.1.0] — 2026-05-21

### added
- Bubble Tea TUI with screen state machine: splash → menu → scenario → sovereign → briefing → game cycles A/B/C
- scenario and sovereign selection screens
- cycle A (world update), B (player commands), C (settlement) turn loop
- `BUILD_AG`, `BUILD_COM`, `BUILD_DEF` commands with CP cost
- economy settlement: grain/gold yield, corruption loss
- Bazi clock: heavenly stems, earthly branches, Wu Xing seasonal elements
- essence drift multiplier applied to effective strategy and CP calculation
- in-memory ledger loaded from `assets/data/*.json` (generated via `make data`)
- YAML master archives: `data/officers.yaml`, `data/cities.yaml`
- 36 Three Kingdoms cities across 8 provinces
- `make tui`, `make test`, `make data` targets
