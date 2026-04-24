# 帝王战纪：三国录 (Sovereign Record: Three Kingdoms Ledger)

> "Sovereignty through the Ledger, Strategy through the Elements."

a headless grand strategy engine and historical simulation set during China's Three Kingdoms era (AD 184–280). powered by a Bazi metaphysical clock and an append-only SQLite ledger.

the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft and elemental alignment. victory is measured by territory, institutional strength, and historical legacy — all recorded in the ledger.

see [SPEC.md](SPEC.md) for full design and architecture documentation.

---

## stack

| layer | tool | role |
| :--- | :--- | :--- |
| engine | Godot 4 | recommended frontend: scenes, tilemap, animation |
| ledger | SQLite | append-only historical record of world state |
| data | YAML | human-readable canonical archives for officers and cities |

---

## quick start

```sh
make          # show available commands
make install  # install dependencies
make run      # launch the game
```

---

## agent context

| module | role |
| :--- | :--- |
| main | entry point: seeds DB from YAML archives, launches frontend |
| engine | three-cycle turn processor (headless, side-effect-free) |
| clock | bazi calendar and essence drift calculation |
| ledger | SQLite layer: schema, CRUD, atomic settle_turn |
| models | typed domain models: Officer, City, Element, Tag, Terrain |
| archive | YAML loader for officers and cities data |
| frontend | Godot scenes: splash, map, turn screen, sovereign dashboard |

data archives (`data/`) are read once at first-run seed and never touched at runtime. all runtime state lives in `ledger.db`.

---

## roadmap

see [SPEC.md](SPEC.md#milestones) for the project roadmap.

---
