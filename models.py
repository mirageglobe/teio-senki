from enum import StrEnum
from dataclasses import dataclass
from pydantic import BaseModel, ConfigDict, Field


class Element(StrEnum):
    WOOD  = "wood"
    FIRE  = "fire"
    EARTH = "earth"
    METAL = "metal"
    WATER = "water"


class Tag(StrEnum):
    # Classification
    CIVIL      = "civil"       # Administrative / fiscal officer
    MILITARY   = "military"    # Combat / campaign officer
    LORD       = "lord"        # Head of state — generates CP; eligible to rule
    # Specialist
    NAVAL      = "naval"       # River / sea combat and logistics
    CAVALRY    = "cavalry"     # Open-terrain mounted charges
    VANGUARD   = "vanguard"    # Leading charges; morale breaking
    SCHOLAR    = "scholar"     # Research, academies, civil output
    DIPLOMAT   = "diplomat"    # Negotiations, alliance management
    ENGINEER   = "engineer"    # Siege, fortification, construction
    MERCHANT   = "merchant"    # Trade route yield and efficiency
    STRATEGIST = "strategist"  # Tactical positioning across all theatres
    LOYALIST   = "loyalist"    # Bribery immunity; loyalty decay resistance
    PIONEER    = "pioneer"     # Frontier governance; new territory integration


@dataclass(frozen=True)
class TagBonus:
    """Situational advantage a tag confers."""
    pillar: str
    multiplier: float
    situation: str


# Pure lookup — single source of truth for tag effects.
TAG_BONUSES: dict[Tag, TagBonus] = {
    Tag.CIVIL:      TagBonus(pillar="governance", multiplier=1.10, situation="Any civil administration or peacetime action"),
    Tag.MILITARY:   TagBonus(pillar="strategy",   multiplier=1.10, situation="Any military operation or campaign"),
    Tag.LORD:       TagBonus(pillar="governance", multiplier=1.20, situation="Any policy action as head of state; also the CP source"),
    Tag.NAVAL:      TagBonus(pillar="strategy",   multiplier=1.25, situation="Naval / river combat and logistics"),
    Tag.CAVALRY:    TagBonus(pillar="valour",     multiplier=1.20, situation="Open-terrain mounted charge"),
    Tag.VANGUARD:   TagBonus(pillar="valour",     multiplier=1.15, situation="Vanguard charge or 1-on-1 duel"),
    Tag.SCHOLAR:    TagBonus(pillar="governance", multiplier=1.20, situation="Academy, research, or civil administration"),
    Tag.DIPLOMAT:   TagBonus(pillar="integrity",  multiplier=1.20, situation="Alliance negotiation or espionage defence"),
    Tag.ENGINEER:   TagBonus(pillar="governance", multiplier=1.20, situation="Siege operation or infrastructure build"),
    Tag.MERCHANT:   TagBonus(pillar="governance", multiplier=1.25, situation="Active trade route"),
    Tag.STRATEGIST: TagBonus(pillar="strategy",   multiplier=1.15, situation="All tactical positioning and logistics"),
    Tag.LOYALIST:   TagBonus(pillar="integrity",  multiplier=1.25, situation="Enemy bribery or plot attempt"),
    Tag.PIONEER:    TagBonus(pillar="governance", multiplier=1.15, situation="Governing a newly annexed territory"),
}


class Terrain(StrEnum):
    PLAIN    = "plain"    # Open ground — cavalry favoured
    MOUNTAIN = "mountain" # High ground — engineers and defence favoured
    FOREST   = "forest"   # Cover — vanguard and ambush favoured
    RIVER    = "river"    # Waterway — naval essential
    COAST    = "coast"    # Sea access — naval dominant
    PASS     = "pass"     # Bottleneck — engineers and defence decisive


@dataclass(frozen=True)
class TerrainBonus:
    """Bonus a terrain type grants to officers with the matching tag."""
    tag: Tag
    multiplier: float
    note: str


# Terrain bonuses applied to officers garrisoning or operating in a city.
TERRAIN_BONUSES: dict[Terrain, list[TerrainBonus]] = {
    Terrain.PLAIN:    [TerrainBonus(Tag.CAVALRY,    1.15, "Open ground favours mounted charges")],
    Terrain.MOUNTAIN: [TerrainBonus(Tag.ENGINEER,   1.20, "High ground amplifies fortification")],
    Terrain.FOREST:   [TerrainBonus(Tag.VANGUARD,   1.10, "Forest cover favours bold infantry")],
    Terrain.RIVER:    [TerrainBonus(Tag.NAVAL,       1.25, "River control is strategically decisive")],
    Terrain.COAST:    [TerrainBonus(Tag.NAVAL,       1.30, "Sea lanes multiply naval power")],
    Terrain.PASS:     [TerrainBonus(Tag.ENGINEER,   1.30, "Passes are natural fortifications"),
                       TerrainBonus(Tag.STRATEGIST,  1.10, "Pass control rewards strategic foresight")],
}


class Officer(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: int | None = None
    name: str
    title: str = ""
    essence: Element
    tags: frozenset[Tag] = frozenset()

    # Five Pillars (1–100)
    strategy: int = Field(ge=1, le=100)    # Army throughput / macro tactics
    valour: int = Field(ge=1, le=100)      # Duel / vanguard charge performance
    governance: int = Field(ge=1, le=100)  # Tax yield / infrastructure compounding
    integrity: int = Field(ge=1, le=100)   # Loyalty resilience / bribery immunity
    loyalty: int = Field(ge=0, le=100, default=50)


class City(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: int | None = None
    name: str           # Romanised name
    chinese: str        # Chinese characters
    region: str         # Province / region
    terrain: Terrain
    x: int = Field(ge=0, le=19)   # Grid column (0 = west, 19 = east)
    y: int = Field(ge=0, le=14)   # Grid row (0 = south, 14 = north)
    population: int = Field(ge=0)
    defense: int = Field(ge=0, le=100)       # Base fortification rating

    # Four Economic Pillars (1–100)
    agriculture: int = Field(ge=1, le=100)   # Wood — Grain output; army upkeep (Blue Chip)
    commerce: int    = Field(ge=1, le=100)   # Metal — Gold flow; recruitment & trade (Liquidity)
    technology: int  = Field(ge=1, le=100)   # Fire — Multiplier on Ag/Com yield (Innovation)
    order: int       = Field(ge=1, le=100)   # Earth — Reduces corruption/riots (The Moat)

    faction: str                             # Starting controller
