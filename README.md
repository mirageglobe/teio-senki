# 帝王战纪：三国录 (Sovereign Record: Three Kingdoms Ledger)

> "Sovereignty through the Ledger, Strategy through the Elements."

a headless grand strategy engine and historical simulation set during China's Three Kingdoms era (AD 184–280). powered by a Bazi metaphysical clock and an in-memory ledger.

the player assumes the role of a sovereign (君主), managing cities and officers through principled statecraft and elemental alignment. victory is measured by territory — all events recorded in the ledger.

see [DESIGN.md](DESIGN.md) for game design, systems, and mechanics. see [SPEC.md](SPEC.md) for architecture and technical documentation.

---

## stack

| layer      | tool                        | role                                              |
| :--------- | :-------------------------- | :------------------------------------------------ |
| language   | TypeScript (strict)         | simulation core, turn loop, ledger                |
| ui         | HTML + Canvas (rot.js)      | browser-native rendering; dumb view over engine   |
| packaging  | Capacitor / Tauri           | mobile (Android, iOS) and desktop wrapping        |
| data       | YAML → JSON                 | human-readable archives converted for runtime     |
| build      | Vite + Vitest               | dev server, production bundle, headless tests     |

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
