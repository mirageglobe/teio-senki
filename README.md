# 帝王战纪：三国录 (Sovereign Record: Three Kingdoms Ledger)

> "Sovereignty through the Ledger, Strategy through the Elements."

a headless grand strategy engine and historical simulation set during China's Three Kingdoms era (AD 184–280). powered by a Bazi metaphysical clock and an in-memory ledger.

the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft and elemental alignment. victory is measured by territory — all events recorded in the ledger.

see [DESIGN.md](DESIGN.md) for game design, systems, and mechanics. see [SPEC.md](SPEC.md) for architecture and technical documentation.

---

## stack

| layer | tool | role |
| :--- | :--- | :--- |
| engine | Go | simulation core, turn loop, ledger |
| scripting | Lua (gopher-lua) | AI behaviours, balance rules, hot-reloadable |
| ui | Bubble Tea (TUI) | terminal frontend; Ebitengine pixel art added at M9 |
| data | YAML → JSON | human-readable archives converted for runtime |

---

## quick start

```sh
make          # show available commands
make install  # install dependencies
make run      # launch the game
```

---

## roadmap

see [SPEC.md](SPEC.md#milestones) for the project roadmap.

---
