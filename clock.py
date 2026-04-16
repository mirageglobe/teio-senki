"""
clock.py — Bazi Metaphysical Clock (八字時鐘).

Tracks the in-game year and month using the traditional Heavenly Stem /
Earthly Branch (天干地支) system. All functions are pure; no side effects.

Game epoch: Month 1, Year 184 AD (甲子年 — Yellow Turban Rebellion).
"""

from dataclasses import dataclass

from models import Element

# ---------------------------------------------------------------------------
# Heavenly Stems 天干 (10-cycle, Yin/Yang pairs of each element)
# ---------------------------------------------------------------------------
STEMS:       tuple[str, ...] = ("甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸")
STEM_PINYIN: tuple[str, ...] = ("jiǎ", "yǐ", "bǐng", "dīng", "wù", "jǐ", "gēng", "xīn", "rén", "guǐ")
STEM_ELEMENT: tuple[Element, ...] = (
    Element.WOOD,  Element.WOOD,
    Element.FIRE,  Element.FIRE,
    Element.EARTH, Element.EARTH,
    Element.METAL, Element.METAL,
    Element.WATER, Element.WATER,
)

# ---------------------------------------------------------------------------
# Earthly Branches 地支 (12-cycle)
# ---------------------------------------------------------------------------
BRANCHES:       tuple[str, ...] = ("子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥")
BRANCH_PINYIN:  tuple[str, ...] = ("zǐ", "chǒu", "yín", "mǎo", "chén", "sì", "wǔ", "wèi", "shēn", "yǒu", "xū", "hài")
BRANCH_ANIMAL:  tuple[str, ...] = ("Rat", "Ox", "Tiger", "Rabbit", "Dragon", "Snake", "Horse", "Goat", "Monkey", "Rooster", "Dog", "Pig")
BRANCH_ELEMENT: tuple[Element, ...] = (
    Element.WATER, Element.EARTH,
    Element.WOOD,  Element.WOOD,  Element.EARTH,
    Element.FIRE,  Element.FIRE,  Element.EARTH,
    Element.METAL, Element.METAL, Element.EARTH,
    Element.WATER,
)

# Month 1 (寅) maps to branch index 2; month n → (n + 1) % 12
_MONTH_BRANCH_OFFSET = 2

# First month stem for each year stem group (甲/己→丙, 乙/庚→戊, 丙/辛→庚, 丁/壬→壬, 戊/癸→甲)
_YEAR_STEM_TO_FIRST_MONTH_STEM: tuple[int, ...] = (2, 4, 6, 8, 0, 2, 4, 6, 8, 0)

# ---------------------------------------------------------------------------
# Wu Xing relationship maps (pure data)
# ---------------------------------------------------------------------------
GENERATING: dict[Element, Element] = {
    Element.WOOD:  Element.FIRE,
    Element.FIRE:  Element.EARTH,
    Element.EARTH: Element.METAL,
    Element.METAL: Element.WATER,
    Element.WATER: Element.WOOD,
}

CONTROLLING: dict[Element, Element] = {
    Element.WOOD:  Element.EARTH,
    Element.EARTH: Element.WATER,
    Element.WATER: Element.FIRE,
    Element.FIRE:  Element.METAL,
    Element.METAL: Element.WOOD,
}

# ---------------------------------------------------------------------------
# Epoch
# ---------------------------------------------------------------------------
EPOCH_YEAR        = 184   # 甲子年 — Yellow Turban Rebellion
EPOCH_STEM_INDEX  = 0     # 甲
EPOCH_BRANCH_INDEX = 0    # 子


# ---------------------------------------------------------------------------
# BaziClock — immutable value object
# ---------------------------------------------------------------------------
@dataclass(frozen=True)
class BaziClock:
    """
    Represents a single point in the in-game Bazi calendar.

    All derived properties are computed from (year, month) with no mutable
    state, making the clock safe to pass around and easy to test.
    """
    year: int   # Absolute in-game year (e.g. 184)
    month: int  # 1–12

    # -- Year stem / branch -------------------------------------------------

    @property
    def year_stem_index(self) -> int:
        return (self.year - EPOCH_YEAR + EPOCH_STEM_INDEX) % 10

    @property
    def year_branch_index(self) -> int:
        return (self.year - EPOCH_YEAR + EPOCH_BRANCH_INDEX) % 12

    @property
    def year_stem(self) -> str:
        return STEMS[self.year_stem_index]

    @property
    def year_branch(self) -> str:
        return BRANCHES[self.year_branch_index]

    @property
    def year_element(self) -> Element:
        return STEM_ELEMENT[self.year_stem_index]

    # -- Month stem / branch ------------------------------------------------

    @property
    def month_stem_index(self) -> int:
        first = _YEAR_STEM_TO_FIRST_MONTH_STEM[self.year_stem_index]
        return (first + self.month - 1) % 10

    @property
    def month_branch_index(self) -> int:
        return (self.month + _MONTH_BRANCH_OFFSET - 1) % 12

    @property
    def month_stem(self) -> str:
        return STEMS[self.month_stem_index]

    @property
    def month_branch(self) -> str:
        return BRANCHES[self.month_branch_index]

    @property
    def month_animal(self) -> str:
        return BRANCH_ANIMAL[self.month_branch_index]

    # -- Derived gameplay values --------------------------------------------

    @property
    def dominant_element(self) -> Element:
        """The ruling Wu Xing element this month (from Earthly Branch)."""
        return BRANCH_ELEMENT[self.month_branch_index]

    @property
    def season(self) -> str:
        return ("Spring", "Spring", "Spring",
                "Summer", "Summer", "Summer",
                "Autumn", "Autumn", "Autumn",
                "Winter", "Winter", "Winter")[self.month - 1]

    @property
    def season_chinese(self) -> str:
        return ("春", "春", "春", "夏", "夏", "夏",
                "秋", "秋", "秋", "冬", "冬", "冬")[self.month - 1]

    @property
    def is_transition_month(self) -> bool:
        """True for months 3, 6, 9, 12 — Earth dominates as seasonal pivot."""
        return self.month % 3 == 0

    # -- Display ------------------------------------------------------------

    @property
    def label(self) -> str:
        """Full Bazi label: e.g. '甲子年 丙寅月'"""
        return f"{self.year_stem}{self.year_branch}年 {self.month_stem}{self.month_branch}月"

    @property
    def summary(self) -> str:
        return (
            f"{self.label}  [{self.season_chinese} · {self.season}]  "
            f"Dominant: {self.dominant_element.value.capitalize()}  "
            f"({BRANCH_ANIMAL[self.month_branch_index]} month)"
        )


# ---------------------------------------------------------------------------
# Pure clock functions
# ---------------------------------------------------------------------------

def make_clock(year: int = EPOCH_YEAR, month: int = 1) -> BaziClock:
    return BaziClock(year=year, month=month)


def advance(clock: BaziClock, months: int = 1) -> BaziClock:
    """Return a new BaziClock advanced by `months` turns."""
    total = (clock.year - EPOCH_YEAR) * 12 + (clock.month - 1) + months
    return BaziClock(year=EPOCH_YEAR + total // 12, month=total % 12 + 1)


def essence_drift(officer_essence: Element, clock: BaziClock) -> float:
    """
    Returns a stat multiplier (0.80–1.25) based on how the officer's Essence
    aligns with the month's dominant Wu Xing element.

    Relationships (generating cycle 生, controlling cycle 克):
      Peak      — essence == dominant              → ×1.25
      Feeding   — essence generates dominant       → ×1.10  (officer feeds the season)
      Nourished — dominant generates essence       → ×1.15  (season lifts the officer)
      Resistant — essence controls dominant        → ×0.90  (officer fights the tide)
      Suppressed— dominant controls essence        → ×0.80  (season bears down on officer)
    """
    d = clock.dominant_element
    e = officer_essence
    if e == d:                      return 1.25
    if GENERATING[e]   == d:        return 1.10
    if GENERATING[d]   == e:        return 1.15
    if CONTROLLING[e]  == d:        return 0.90
    if CONTROLLING[d]  == e:        return 0.80
    return 1.0   # structural fallback; all 5-element pairs are covered above


def effective_stat(base: int, drift: float) -> int:
    """Apply an essence drift multiplier to a base stat, clamped to 1–100."""
    return max(1, min(100, round(base * drift)))
