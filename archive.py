"""
archive.py — YAML Master Archive loader.

Reads human-readable YAML config and returns validated, immutable Pydantic
models. This is the only module that touches PyYAML; everything else speaks
in domain value objects.
"""

from pathlib import Path
from typing import Any

import yaml

from models import City, Element, Officer, Tag, Terrain

OFFICERS_YAML = Path(__file__).parent / "data" / "officers.yaml"
CITIES_YAML   = Path(__file__).parent / "data" / "cities.yaml"


# ---------------------------------------------------------------------------
# Pure transformers
# ---------------------------------------------------------------------------

def _parse_tags(raw: list[str]) -> frozenset[Tag]:
    return frozenset(Tag(t.lower()) for t in raw)


def _parse_officer(record: dict[str, Any]) -> Officer:
    return Officer(
        name=record["name"],
        title=record.get("title", ""),
        essence=Element(record["essence"].lower()),
        tags=_parse_tags(record.get("tags", [])),
        strategy=record["strategy"],
        valour=record["valour"],
        governance=record["governance"],
        integrity=record["integrity"],
        loyalty=record.get("loyalty", 50),
    )


def _parse_city(record: dict[str, Any]) -> City:
    return City(
        name=record["name"],
        chinese=record["chinese"],
        region=record["region"],
        terrain=Terrain(record["terrain"].lower()),
        x=record["x"],
        y=record["y"],
        population=record["population"],
        defense=record["defense"],
        agriculture=record["agriculture"],
        commerce=record["commerce"],
        technology=record["technology"],
        order=record["order"],
        faction=record["faction"],
    )


# ---------------------------------------------------------------------------
# Loaders
# ---------------------------------------------------------------------------

def load_officers(path: Path = OFFICERS_YAML) -> tuple[Officer, ...]:
    """Parse the YAML archive and return a validated, immutable officer roster."""
    raw = yaml.safe_load(path.read_text(encoding="utf-8"))
    return tuple(map(_parse_officer, raw["officers"]))


def load_cities(path: Path = CITIES_YAML) -> tuple[City, ...]:
    """Parse the YAML archive and return a validated, immutable city roster."""
    raw = yaml.safe_load(path.read_text(encoding="utf-8"))
    return tuple(map(_parse_city, raw["cities"]))
