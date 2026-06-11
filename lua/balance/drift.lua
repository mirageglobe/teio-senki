-- drift.lua: Wu Xing drift multiplier overrides.
-- Edit and reload with 'make reload' during development to tune live.
-- Go engine reads these values via the bridge; they do NOT override core formulas directly.

drift = {
  PEAK       = 1.25,
  NOURISHED  = 1.15,
  FEEDING    = 1.10,
  RESISTANT  = 0.90,
  SUPPRESSED = 0.80,
}
