"""
engine.py — Turn Engine: three-phase batch processor.

Each turn is an atomic batch job:
  Phase A — Market & Metaphysic Shift  (world updates, read-only on DB)
  Phase B — Resource Allocation         (player issues Commands within CP budget)
  Phase C — Settlement                  (engine resolves deltas, commits to Ledger)

All phase functions are pure. Side effects (DB writes) are confined to
ledger.settle_turn(), which wraps Phase C in a single SQLite transaction.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import StrEnum
from typing import Any

from clock import BaziClock, advance
from models import City, Element


# ---------------------------------------------------------------------------
# Phase A — World Events
# ---------------------------------------------------------------------------

class EventType(StrEnum):
    SEASON_SHIFT  = "season_shift"   # Clock advanced; new element dominant
    GRAIN_SIGNAL  = "grain_signal"   # Seasonal agriculture price pressure
    GOLD_SIGNAL   = "gold_signal"    # Commerce velocity signal
    TECH_PULSE    = "tech_pulse"     # Technology multiplier shift
    ORDER_DRIFT   = "order_drift"    # Social stability signal


@dataclass(frozen=True)
class WorldEvent:
    event_type: EventType
    description: str
    effects: dict[str, Any]   # declarative delta map — no mutation here


# Seasonal effect on each economic pillar (delta per city, applied in Phase C)
_SEASON_DELTAS: dict[Element, dict[str, int]] = {
    #                    AG   COM  TECH  ORD
    Element.WOOD:  dict(agriculture=+2, commerce= 0, technology=+1, order= 0),  # Spring growth
    Element.FIRE:  dict(agriculture=+1, commerce=+2, technology=+2, order=-1),  # Summer heat / high activity
    Element.EARTH: dict(agriculture= 0, commerce= 0, technology= 0, order=+2),  # Transition — stability
    Element.METAL: dict(agriculture=+3, commerce=+1, technology= 0, order=+1),  # Autumn harvest
    Element.WATER: dict(agriculture=-2, commerce=-1, technology= 0, order= 0),  # Winter contraction
}


def _season_events(clock: BaziClock) -> tuple[WorldEvent, ...]:
    dom  = clock.dominant_element
    d    = _SEASON_DELTAS[dom]
    dom_name = dom.value.capitalize()

    shift = WorldEvent(
        event_type=EventType.SEASON_SHIFT,
        description=(
            f"{clock.label} — {clock.season} ({clock.season_chinese}). "
            f"{dom_name} element dominant. {clock.month_animal} month."
        ),
        effects={"dominant_element": dom.value, "season": clock.season},
    )

    grain = WorldEvent(
        event_type=EventType.GRAIN_SIGNAL,
        description=(
            f"Agriculture {'+' if d['agriculture'] >= 0 else ''}{d['agriculture']} "
            f"— {clock.season} conditions ({dom_name} phase)."
        ),
        effects={"agriculture_delta": d["agriculture"]},
    )

    gold = WorldEvent(
        event_type=EventType.GOLD_SIGNAL,
        description=(
            f"Commerce {'+' if d['commerce'] >= 0 else ''}{d['commerce']} "
            f"— market velocity in {clock.season}."
        ),
        effects={"commerce_delta": d["commerce"]},
    )

    tech = WorldEvent(
        event_type=EventType.TECH_PULSE,
        description=(
            f"Technology {'+' if d['technology'] >= 0 else ''}{d['technology']} "
            f"— {dom_name} drives {'innovation' if d['technology'] >= 0 else 'stagnation'}."
        ),
        effects={"technology_delta": d["technology"]},
    )

    order_ev = WorldEvent(
        event_type=EventType.ORDER_DRIFT,
        description=(
            f"Order {'+' if d['order'] >= 0 else ''}{d['order']} "
            f"— {'stability consolidates' if d['order'] >= 0 else 'unrest stirs'} in {clock.season}."
        ),
        effects={"order_delta": d["order"]},
    )

    return (shift, grain, gold, tech, order_ev)


def phase_a(clock: BaziClock) -> tuple[BaziClock, tuple[WorldEvent, ...]]:
    """
    Advance the Bazi clock by one month and generate the world event stream.
    Returns (new_clock, events). Pure — no DB access.
    """
    new_clock = advance(clock)
    return new_clock, _season_events(new_clock)


# ---------------------------------------------------------------------------
# Phase B — Commands
# ---------------------------------------------------------------------------

class CommandType(StrEnum):
    BUILD_AGRICULTURE = "build_agriculture"  # +2 AG in target city
    BUILD_COMMERCE    = "build_commerce"     # +2 COM in target city
    BUILD_TECHNOLOGY  = "build_technology"   # +2 TECH in target city
    BUILD_ORDER       = "build_order"        # +2 ORD in target city
    BUILD_DEFENSE     = "build_defense"      # +2 DEF in target city


# Cost in Command Points per command type
COMMAND_COSTS: dict[CommandType, int] = {
    CommandType.BUILD_AGRICULTURE: 1,
    CommandType.BUILD_COMMERCE:    1,
    CommandType.BUILD_TECHNOLOGY:  2,
    CommandType.BUILD_ORDER:       1,
    CommandType.BUILD_DEFENSE:     2,
}

# Stat delta each command applies to the target city
COMMAND_DELTAS: dict[CommandType, dict[str, int]] = {
    CommandType.BUILD_AGRICULTURE: {"agriculture": 2},
    CommandType.BUILD_COMMERCE:    {"commerce":    2},
    CommandType.BUILD_TECHNOLOGY:  {"technology":  2},
    CommandType.BUILD_ORDER:       {"order":       2},
    CommandType.BUILD_DEFENSE:     {"defense":     2},
}


@dataclass(frozen=True)
class Command:
    type: CommandType
    city_id: int
    cost: int = field(init=False)

    def __post_init__(self) -> None:
        object.__setattr__(self, "cost", COMMAND_COSTS[self.type])


def command_points(ruler_strategy: int) -> int:
    """Available CP this turn = strategy ÷ 10, minimum 5.
    Pass the active lord's strategy stat. Only officers with Tag.LORD
    are eligible to be set as the active ruler."""
    return max(5, ruler_strategy // 10)


def find_lords(officers) -> list:
    """Return officers carrying the lord tag, sorted by strategy descending."""
    from models import Tag
    return sorted(
        (o for o in officers if Tag.LORD in o.tags),
        key=lambda o: o.strategy,
        reverse=True,
    )


def validate_commands(
    commands: tuple[Command, ...],
    available_cp: int,
) -> tuple[Command, ...]:
    """
    Greedy CP filter — approve commands in order until the budget is exhausted.
    Returns the approved subset (pure, no mutation).
    """
    used, approved = 0, []
    for cmd in commands:
        if used + cmd.cost <= available_cp:
            approved.append(cmd)
            used += cmd.cost
    return tuple(approved)


# ---------------------------------------------------------------------------
# Phase C — Resolution (pure delta computation)
# ---------------------------------------------------------------------------

def _clamp(value: int, lo: int = 1, hi: int = 100) -> int:
    return max(lo, min(hi, value))


def apply_seasonal_delta(city: City, events: tuple[WorldEvent, ...]) -> City:
    """
    Pure: fold all world-event deltas onto a City, return updated City.
    Clamps each pillar to [1, 100].
    """
    merged: dict[str, int] = {}
    for ev in events:
        for key, val in ev.effects.items():
            if key.endswith("_delta"):
                pillar = key.removesuffix("_delta")
                merged[pillar] = merged.get(pillar, 0) + val

    updates = {
        pillar: _clamp(getattr(city, pillar) + delta)
        for pillar, delta in merged.items()
        if hasattr(city, pillar)
    }
    return city.model_copy(update=updates) if updates else city


def apply_command(city: City, command: Command) -> City:
    """Pure: apply one Command's delta to a City."""
    delta = COMMAND_DELTAS[command.type]
    updates = {
        pillar: _clamp(getattr(city, pillar) + d)
        for pillar, d in delta.items()
    }
    return city.model_copy(update=updates)


def resolve_turn(
    cities: dict[int, City],
    events: tuple[WorldEvent, ...],
    commands: tuple[Command, ...],
) -> tuple[dict[int, City], tuple[WorldEvent, ...]]:
    """
    Pure resolution: apply seasonal deltas to all cities, then apply commands.
    Returns (updated_cities, command_events).
    """
    # Apply seasonal deltas globally
    updated = {cid: apply_seasonal_delta(c, events) for cid, c in cities.items()}

    # Apply commands to their target cities
    cmd_events: list[WorldEvent] = []
    for cmd in commands:
        if cmd.city_id in updated:
            before = updated[cmd.city_id]
            after  = apply_command(before, cmd)
            updated[cmd.city_id] = after
            cmd_events.append(WorldEvent(
                event_type=EventType.SEASON_SHIFT,   # reuse as generic log entry
                description=f"Command {cmd.type.value} on city {cmd.city_id}: {COMMAND_DELTAS[cmd.type]}",
                effects={"city_id": cmd.city_id, "command": cmd.type.value,
                         "delta": COMMAND_DELTAS[cmd.type]},
            ))

    return updated, tuple(cmd_events)
