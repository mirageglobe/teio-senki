"""
app.py — Textual TUI for 帝王战纪：三国录

Uses the alternate-screen buffer via Textual's Driver. The terminal is
restored cleanly on any exit path — Q, Ctrl+C, or exception.
"""
from __future__ import annotations

import sqlite3

from rich.text import Text
from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.screen import Screen
from textual.widgets import (
    DataTable, Footer, Header, Input, RichLog, Static,
    TabbedContent, TabPane,
)

from clock import BaziClock, essence_drift, effective_stat
from engine import (
    Command, CommandType,
    command_points, validate_commands,
    phase_a, resolve_turn, find_lords,
)
from ledger import (
    get_all_cities, get_all_officers, get_clock,
    get_ledger_log, settle_turn as db_settle,
)
from models import Element, Tag


# ---------------------------------------------------------------------------
# Rich helpers
# ---------------------------------------------------------------------------

_ELEM_STYLE: dict[Element, str] = {
    Element.WOOD:  "green",
    Element.FIRE:  "bold red",
    Element.EARTH: "yellow",
    Element.METAL: "white",
    Element.WATER: "cyan",
}

_CMD_MAP: dict[str, CommandType] = {
    "ag":   CommandType.BUILD_AGRICULTURE,
    "com":  CommandType.BUILD_COMMERCE,
    "tech": CommandType.BUILD_TECHNOLOGY,
    "ord":  CommandType.BUILD_ORDER,
    "def":  CommandType.BUILD_DEFENSE,
}


def _elem(e: Element) -> Text:
    return Text(e.value, style=_ELEM_STYLE[e])


def _drift(d: float) -> Text:
    if d > 1.0:
        return Text(f"+{d:.2f}", style="bold green")
    if d < 1.0:
        return Text(f" {d:.2f}", style="bold red")
    return Text(f" {d:.2f}", style="dim")


# ---------------------------------------------------------------------------
# Turn Screen
# ---------------------------------------------------------------------------

class TurnScreen(Screen):
    """
    Three-phase turn execution pushed onto the screen stack.
    Phase A runs automatically on mount.
    Phase B accepts player commands via Input.
    Phase C settles on 'done' / empty input / S key.
    Dismisses itself when the player acknowledges results.
    """

    TITLE = "帝王战纪 — End Turn"

    CSS = """
    TurnScreen { background: $background; }
    #turn-log  { height: 1fr; border: solid $primary; margin: 0 1; }
    #q-table   { height: 7; border: solid $secondary; margin: 0 1; }
    #cp-bar    { height: 1; padding: 0 2; color: $text-muted; }
    #cmd-input { margin: 0 1 1 1; }
    """

    BINDINGS = [
        Binding("s",      "settle", "Settle Turn", priority=True),
        Binding("escape", "cancel", "Cancel / Back"),
    ]

    def __init__(self, conn: sqlite3.Connection) -> None:
        super().__init__()
        self.conn = conn
        self._commands: list[Command] = []
        self._cp_total: int = 5
        self._cp_used: int = 0
        self._new_clock: BaziClock | None = None
        self._a_events: tuple = ()
        self._settled = False

    # -- Layout --------------------------------------------------------------

    def compose(self) -> ComposeResult:
        yield Header()
        yield RichLog(id="turn-log", highlight=True, markup=True, wrap=True)
        yield DataTable(id="q-table", cursor_type="none", zebra_stripes=True)
        yield Static("", id="cp-bar")
        yield Input(
            id="cmd-input",
            placeholder="ag·com·tech·ord·def <city_id>   |   [S] settle   |   [Esc] cancel",
        )
        yield Footer()

    def on_mount(self) -> None:
        qt = self.query_one("#q-table", DataTable)
        qt.add_columns("Command", "City ID", "CP Cost")

        self._run_phase_a()
        self.query_one("#cmd-input", Input).focus()

    # -- Phase A -------------------------------------------------------------

    def _run_phase_a(self) -> None:
        log = self.query_one("#turn-log", RichLog)
        clock = get_clock(self.conn)
        self._new_clock, self._a_events = phase_a(clock)

        log.write("[bold cyan]── PHASE A  Market & Metaphysic Shift ──────────────[/]")
        log.write(
            f"  {clock.label}  [dim]→[/]  [bold]{self._new_clock.label}[/]"
            f"  {self._new_clock.season_chinese} {self._new_clock.season}"
            f"  ·  [bold]{self._new_clock.dominant_element.value.capitalize()}[/] dominant\n"
        )
        for ev in self._a_events:
            log.write(f"  [dim]{ev.event_type.value:<16}[/]  {ev.description}")

        # Active lord → CP
        lords = find_lords(get_all_officers(self.conn))
        log.write("\n[bold cyan]── PHASE B  Resource Allocation ─────────────────────[/]")
        if lords:
            lo = lords[0]
            drift = essence_drift(lo.essence, self._new_clock)
            eff   = effective_stat(lo.strategy, drift)
            self._cp_total = command_points(eff)
            style = _ELEM_STYLE[lo.essence]
            log.write(
                f"  Active lord: [bold]{lo.name}[/]"
                f"  essence=[{style}]{lo.essence.value}[/]"
                f"  drift {drift:.2f}"
                f"  STR {eff}"
                f"  →  [bold yellow]{self._cp_total} CP[/]"
            )
        else:
            log.write("  No lord found — default 5 CP.")

        self._update_cp_bar()

    # -- Phase B helpers -----------------------------------------------------

    def _update_cp_bar(self) -> None:
        left = self._cp_total - self._cp_used
        self.query_one("#cp-bar", Static).update(
            f"  CP {self._cp_used}/{self._cp_total}  ·  {left} remaining"
            f"  ·  {len(self._commands)} command(s) queued"
        )

    def on_input_submitted(self, event: Input.Submitted) -> None:
        raw = event.value.strip().lower()
        event.input.clear()

        # Empty input after settlement → dismiss
        if self._settled:
            self.dismiss()
            return

        # Empty input or explicit "done" → settle
        if not raw or raw in ("done", "s"):
            self.action_settle()
            return

        parts = raw.split()
        log = self.query_one("#turn-log", RichLog)

        if len(parts) != 2 or parts[0] not in _CMD_MAP:
            log.write(
                f"  [red]Unknown.[/] Valid: "
                f"[bold]ag com tech ord def[/] <city_id>"
            )
            return

        try:
            city_id = int(parts[1])
        except ValueError:
            log.write("  [red]City ID must be an integer.[/]")
            return

        cmd  = Command(_CMD_MAP[parts[0]], city_id=city_id)
        left = self._cp_total - self._cp_used
        if cmd.cost > left:
            log.write(f"  [red]Not enough CP[/] — need {cmd.cost}, have {left}.")
            return

        self._commands.append(cmd)
        self._cp_used += cmd.cost
        self.query_one("#q-table", DataTable).add_row(
            cmd.type.value, str(city_id), str(cmd.cost)
        )
        log.write(
            f"  [green]Queued[/] {cmd.type.value}"
            f" → city {city_id}  (cost {cmd.cost} CP)"
        )
        self._update_cp_bar()

    # -- Phase C -------------------------------------------------------------

    def action_settle(self) -> None:
        if self._settled:
            self.dismiss()
            return

        self._settled = True
        log   = self.query_one("#turn-log", RichLog)
        inp   = self.query_one("#cmd-input", Input)

        log.write("\n[bold cyan]── PHASE C  Settlement ──────────────────────────────[/]")

        approved        = validate_commands(tuple(self._commands), self._cp_total)
        cities_by_id    = {c.id: c for c in get_all_cities(self.conn) if c.id is not None}
        updated, c_evts = resolve_turn(cities_by_id, self._a_events, approved)

        db_settle(self.conn, self._new_clock, updated, self._a_events, c_evts)

        changed = 0
        for cid, after in updated.items():
            before  = cities_by_id[cid]
            deltas  = {
                f: getattr(after, f) - getattr(before, f)
                for f in ("agriculture", "commerce", "technology", "order", "defense")
                if getattr(after, f) != getattr(before, f)
            }
            if deltas:
                parts = "  ".join(
                    f"[{'green' if v > 0 else 'red'}]{k[:4].upper()} "
                    f"{'+' if v > 0 else ''}{v}[/]"
                    for k, v in deltas.items()
                )
                log.write(f"  {after.name:<14} {parts}")
                changed += 1

        log.write(
            f"\n  [bold green]Committed.[/]  {self._new_clock.label}"
            f"  ·  {changed} cities updated."
            f"\n\n  [dim]Press [bold]Enter[/bold] or [bold]Esc[/bold] to return.[/]"
        )
        inp.placeholder = "Press Enter to return…"
        inp.focus()

    def action_cancel(self) -> None:
        self.dismiss()


# ---------------------------------------------------------------------------
# Main App
# ---------------------------------------------------------------------------

class SovereignApp(App):
    TITLE = "帝王战纪：三国录"

    CSS = """
    Screen       { background: $background; }
    DataTable    { height: 1fr; }
    TabbedContent{ height: 1fr; }
    TabPane      { padding: 0; }
    #log-panel   { height: 1fr; padding: 0 1; }
    """

    BINDINGS = [
        Binding("t",     "end_turn",       "End Turn",  priority=True),
        Binding("1",     "switch_tab('tab-officers')", "Officers"),
        Binding("2",     "switch_tab('tab-cities')",   "Cities"),
        Binding("3",     "switch_tab('tab-log')",      "Log"),
        Binding("r",     "refresh_all",    "Refresh"),
        Binding("q",     "quit",           "Exit"),
    ]

    def __init__(self, conn: sqlite3.Connection) -> None:
        super().__init__()
        self.conn = conn

    # -- Layout --------------------------------------------------------------

    def compose(self) -> ComposeResult:
        yield Header()
        with TabbedContent(id="main-tabs"):
            with TabPane("Officers [1]", id="tab-officers"):
                yield DataTable(
                    id="officers-table", cursor_type="row", zebra_stripes=True
                )
            with TabPane("Cities [2]", id="tab-cities"):
                yield DataTable(
                    id="cities-table", cursor_type="row", zebra_stripes=True
                )
            with TabPane("Log [3]", id="tab-log"):
                yield RichLog(id="log-panel", highlight=True, markup=True, wrap=True)
        yield Footer()

    def on_mount(self) -> None:
        self._populate_officers()
        self._populate_cities()
        self._sync_subtitle()

    # -- Data population -----------------------------------------------------

    def _populate_officers(self) -> None:
        table = self.query_one("#officers-table", DataTable)
        table.clear(columns=True)
        table.add_columns(
            "ID", "Name", "Title", "Essence", "Drift",
            "STR*", "VAL*", "GOV*", "INT*", "LOY", "Tags",
        )
        clock = get_clock(self.conn)
        for o in get_all_officers(self.conn):
            d    = essence_drift(o.essence, clock)
            tags = " ".join(sorted(t.value for t in o.tags))
            name = Text(o.name, style="bold gold1" if Tag.LORD in o.tags else "")
            table.add_row(
                str(o.id), name, o.title, _elem(o.essence), _drift(d),
                str(effective_stat(o.strategy,   d)),
                str(effective_stat(o.valour,     d)),
                str(effective_stat(o.governance, d)),
                str(effective_stat(o.integrity,  d)),
                str(o.loyalty),
                Text(tags, style="dim"),
            )

    def _populate_cities(self) -> None:
        table = self.query_one("#cities-table", DataTable)
        table.clear(columns=True)
        table.add_columns(
            "ID", "Name", "Chinese", "Region", "Terrain",
            "AG", "COM", "TECH", "ORD", "DEF", "Pop", "Faction",
        )
        for c in get_all_cities(self.conn):
            table.add_row(
                str(c.id), c.name, c.chinese, c.region, c.terrain.value,
                str(c.agriculture), str(c.commerce),
                str(c.technology),  str(c.order), str(c.defense),
                f"{c.population:,}", c.faction,
            )

    def _populate_log(self) -> None:
        log = self.query_one("#log-panel", RichLog)
        log.clear()
        for row in reversed(get_ledger_log(self.conn, limit=100)):
            log.write(
                f"[dim]{row['year']}年M{row['month']:02d} Ph{row['phase']}[/]"
                f"  [cyan]{row['event_type']:<16}[/]  {row['description']}"
            )

    def _sync_subtitle(self) -> None:
        self.sub_title = get_clock(self.conn).summary

    # -- Actions -------------------------------------------------------------

    def action_end_turn(self) -> None:
        def _after_turn(result=None) -> None:
            self._populate_officers()
            self._populate_cities()
            self._populate_log()
            self._sync_subtitle()

        self.push_screen(TurnScreen(self.conn), callback=_after_turn)

    def action_switch_tab(self, tab_id: str) -> None:
        tabs = self.query_one("#main-tabs", TabbedContent)
        tabs.active = tab_id
        if tab_id == "tab-log":
            self._populate_log()

    def action_refresh_all(self) -> None:
        self._populate_officers()
        self._populate_cities()
        self._populate_log()
        self._sync_subtitle()
