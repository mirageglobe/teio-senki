"""
ledger.py — SQLite persistence layer for the Sovereign Record.

All functions are pure data-in / data-out. Side effects (DB writes) are
explicit and isolated here; nothing else in the engine touches sqlite3 directly.
"""

import json
import sqlite3
from pathlib import Path

from clock import BaziClock, make_clock
from models import City, Element, Officer, Tag, Terrain

DB_PATH = Path("ledger.db")


# ---------------------------------------------------------------------------
# Connection
# ---------------------------------------------------------------------------

def connect(path: Path = DB_PATH) -> sqlite3.Connection:
    conn = sqlite3.connect(path)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("PRAGMA foreign_keys=ON")
    return conn


# ---------------------------------------------------------------------------
# Schema
# ---------------------------------------------------------------------------

_SCHEMA = """
CREATE TABLE IF NOT EXISTS officers (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    title       TEXT    NOT NULL DEFAULT '',
    essence     TEXT    NOT NULL CHECK(essence IN ('wood', 'fire', 'earth', 'metal', 'water')),
    strategy    INTEGER NOT NULL CHECK(strategy   BETWEEN 1 AND 100),
    valour      INTEGER NOT NULL CHECK(valour     BETWEEN 1 AND 100),
    governance  INTEGER NOT NULL CHECK(governance BETWEEN 1 AND 100),
    integrity   INTEGER NOT NULL CHECK(integrity  BETWEEN 1 AND 100),
    loyalty     INTEGER NOT NULL DEFAULT 50 CHECK(loyalty BETWEEN 0 AND 100)
);

CREATE TABLE IF NOT EXISTS officer_tags (
    officer_id  INTEGER NOT NULL REFERENCES officers(id) ON DELETE CASCADE,
    tag         TEXT    NOT NULL,
    PRIMARY KEY (officer_id, tag)
);

CREATE TABLE IF NOT EXISTS cities (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    chinese     TEXT    NOT NULL,
    region      TEXT    NOT NULL,
    terrain     TEXT    NOT NULL CHECK(terrain IN ('plain', 'mountain', 'forest', 'river', 'coast', 'pass')),
    x           INTEGER NOT NULL CHECK(x BETWEEN 0 AND 19),
    y           INTEGER NOT NULL CHECK(y BETWEEN 0 AND 14),
    population  INTEGER NOT NULL CHECK(population >= 0),
    defense     INTEGER NOT NULL CHECK(defense     BETWEEN 1 AND 100),
    agriculture INTEGER NOT NULL CHECK(agriculture BETWEEN 1 AND 100),
    commerce    INTEGER NOT NULL CHECK(commerce    BETWEEN 1 AND 100),
    technology  INTEGER NOT NULL CHECK(technology  BETWEEN 1 AND 100),
    "order"     INTEGER NOT NULL CHECK("order"     BETWEEN 1 AND 100),
    faction     TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS game_clock (
    id      INTEGER PRIMARY KEY CHECK(id = 1),
    year    INTEGER NOT NULL,
    month   INTEGER NOT NULL CHECK(month BETWEEN 1 AND 12)
);

CREATE TABLE IF NOT EXISTS ledger_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    year        INTEGER NOT NULL,
    month       INTEGER NOT NULL,
    phase       TEXT    NOT NULL CHECK(phase IN ('A', 'B', 'C')),
    event_type  TEXT    NOT NULL,
    description TEXT    NOT NULL,
    effects_json TEXT
);
"""


def init_db(conn: sqlite3.Connection) -> None:
    conn.executescript(_SCHEMA)
    conn.commit()


# ---------------------------------------------------------------------------
# Row mapping (pure)
# ---------------------------------------------------------------------------

def _tags_for(conn: sqlite3.Connection, officer_id: int) -> frozenset[Tag]:
    rows = conn.execute(
        "SELECT tag FROM officer_tags WHERE officer_id = ?", (officer_id,)
    ).fetchall()
    return frozenset(Tag(r["tag"]) for r in rows)


def _row_to_officer(conn: sqlite3.Connection, row: sqlite3.Row) -> Officer:
    return Officer(
        id=row["id"],
        name=row["name"],
        title=row["title"],
        essence=Element(row["essence"]),
        tags=_tags_for(conn, row["id"]),
        strategy=row["strategy"],
        valour=row["valour"],
        governance=row["governance"],
        integrity=row["integrity"],
        loyalty=row["loyalty"],
    )


def _row_to_city(row: sqlite3.Row) -> City:
    return City(
        id=row["id"],
        name=row["name"],
        chinese=row["chinese"],
        region=row["region"],
        terrain=Terrain(row["terrain"]),
        x=row["x"],
        y=row["y"],
        population=row["population"],
        defense=row["defense"],
        agriculture=row["agriculture"],
        commerce=row["commerce"],
        technology=row["technology"],
        order=row["order"],
        faction=row["faction"],
    )


# ---------------------------------------------------------------------------
# Officer writes / reads
# ---------------------------------------------------------------------------

def _insert_tags(conn: sqlite3.Connection, officer_id: int, tags: frozenset[Tag]) -> None:
    conn.executemany(
        "INSERT OR IGNORE INTO officer_tags (officer_id, tag) VALUES (?, ?)",
        ((officer_id, tag.value) for tag in tags),
    )


def add_officer(conn: sqlite3.Connection, officer: Officer) -> Officer:
    cursor = conn.execute(
        """
        INSERT INTO officers (name, title, essence, strategy, valour, governance, integrity, loyalty)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (officer.name, officer.title, officer.essence.value,
         officer.strategy, officer.valour, officer.governance,
         officer.integrity, officer.loyalty),
    )
    assigned_id: int = cursor.lastrowid  # type: ignore[assignment]
    _insert_tags(conn, assigned_id, officer.tags)
    conn.commit()
    return officer.model_copy(update={"id": assigned_id})


def get_all_officers(conn: sqlite3.Connection) -> list[Officer]:
    rows = conn.execute("SELECT * FROM officers ORDER BY name").fetchall()
    return [_row_to_officer(conn, row) for row in rows]


def get_officer_by_id(conn: sqlite3.Connection, officer_id: int) -> Officer | None:
    row = conn.execute("SELECT * FROM officers WHERE id = ?", (officer_id,)).fetchone()
    return _row_to_officer(conn, row) if row else None


def find_officers_by_tag(conn: sqlite3.Connection, tag: Tag) -> list[Officer]:
    rows = conn.execute(
        """
        SELECT o.* FROM officers o
        JOIN officer_tags t ON o.id = t.officer_id
        WHERE t.tag = ?
        ORDER BY o.name
        """,
        (tag.value,),
    ).fetchall()
    return [_row_to_officer(conn, row) for row in rows]


def find_officers_by_essence(conn: sqlite3.Connection, essence: Element) -> list[Officer]:
    rows = conn.execute(
        "SELECT * FROM officers WHERE essence = ? ORDER BY name", (essence.value,)
    ).fetchall()
    return [_row_to_officer(conn, row) for row in rows]


# ---------------------------------------------------------------------------
# City writes / reads
# ---------------------------------------------------------------------------

def add_city(conn: sqlite3.Connection, city: City) -> City:
    cursor = conn.execute(
        """
        INSERT INTO cities
            (name, chinese, region, terrain, x, y, population, defense,
             agriculture, commerce, technology, "order", faction)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (city.name, city.chinese, city.region, city.terrain.value,
         city.x, city.y, city.population, city.defense,
         city.agriculture, city.commerce, city.technology, city.order, city.faction),
    )
    conn.commit()
    return city.model_copy(update={"id": cursor.lastrowid})


def get_all_cities(conn: sqlite3.Connection) -> list[City]:
    rows = conn.execute("SELECT * FROM cities ORDER BY name").fetchall()
    return list(map(_row_to_city, rows))


def get_city_by_id(conn: sqlite3.Connection, city_id: int) -> City | None:
    row = conn.execute("SELECT * FROM cities WHERE id = ?", (city_id,)).fetchone()
    return _row_to_city(row) if row else None


def find_cities_by_faction(conn: sqlite3.Connection, faction: str) -> list[City]:
    rows = conn.execute(
        "SELECT * FROM cities WHERE faction = ? ORDER BY name", (faction,)
    ).fetchall()
    return list(map(_row_to_city, rows))


def find_cities_by_terrain(conn: sqlite3.Connection, terrain: Terrain) -> list[City]:
    rows = conn.execute(
        "SELECT * FROM cities WHERE terrain = ? ORDER BY name", (terrain.value,)
    ).fetchall()
    return list(map(_row_to_city, rows))


# ---------------------------------------------------------------------------
# Game clock reads / writes
# ---------------------------------------------------------------------------

def get_clock(conn: sqlite3.Connection) -> BaziClock:
    row = conn.execute("SELECT year, month FROM game_clock WHERE id = 1").fetchone()
    if row:
        return BaziClock(year=row["year"], month=row["month"])
    return make_clock()


def set_clock(conn: sqlite3.Connection, clock: BaziClock) -> None:
    conn.execute(
        """
        INSERT INTO game_clock (id, year, month) VALUES (1, ?, ?)
        ON CONFLICT(id) DO UPDATE SET year = excluded.year, month = excluded.month
        """,
        (clock.year, clock.month),
    )
    conn.commit()


# ---------------------------------------------------------------------------
# Ledger log
# ---------------------------------------------------------------------------

def _log_events(
    conn: sqlite3.Connection,
    clock: BaziClock,
    phase: str,
    events,
) -> None:
    conn.executemany(
        """
        INSERT INTO ledger_log (year, month, phase, event_type, description, effects_json)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        (
            (clock.year, clock.month, phase,
             ev.event_type.value, ev.description,
             json.dumps(ev.effects))
            for ev in events
        ),
    )


def get_ledger_log(conn: sqlite3.Connection, limit: int = 20) -> list[sqlite3.Row]:
    return conn.execute(
        "SELECT * FROM ledger_log ORDER BY id DESC LIMIT ?", (limit,)
    ).fetchall()


# ---------------------------------------------------------------------------
# Turn settlement — single atomic transaction
# ---------------------------------------------------------------------------

def settle_turn(
    conn: sqlite3.Connection,
    new_clock: BaziClock,
    updated_cities: dict[int, City],
    phase_a_events,
    phase_c_events,
) -> None:
    """
    Commit a resolved turn to the Ledger in one atomic transaction:
      1. Advance the game clock.
      2. Persist updated city economic pillars.
      3. Append Phase A world events to ledger_log.
      4. Append Phase C command events to ledger_log.
    """
    with conn:
        # 1. Clock
        conn.execute(
            """
            INSERT INTO game_clock (id, year, month) VALUES (1, ?, ?)
            ON CONFLICT(id) DO UPDATE SET year = excluded.year, month = excluded.month
            """,
            (new_clock.year, new_clock.month),
        )

        # 2. Cities — update only the four economic pillars + defense
        for city in updated_cities.values():
            conn.execute(
                """
                UPDATE cities
                SET agriculture = ?, commerce = ?, technology = ?, "order" = ?, defense = ?
                WHERE id = ?
                """,
                (city.agriculture, city.commerce, city.technology,
                 city.order, city.defense, city.id),
            )

        # 3. Phase A events
        _log_events(conn, new_clock, "A", phase_a_events)

        # 4. Phase C command events
        _log_events(conn, new_clock, "C", phase_c_events)
