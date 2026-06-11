// Package essence owns Wu Xing elemental drift calculations.
// It does NOT own officer state or clock state.
package essence

import "math"

// generation cycle: Wood→Fire→Earth→Metal→Water→Wood
var generation = map[string]string{
	"Wood": "Fire", "Fire": "Earth", "Earth": "Metal", "Metal": "Water", "Water": "Wood",
}

// control cycle: Wood→Earth→Water→Fire→Metal→Wood
var control = map[string]string{
	"Wood": "Earth", "Earth": "Water", "Water": "Fire", "Fire": "Metal", "Metal": "Wood",
}

// DriftMultiplier returns the seasonal influence on an officer's effective stats.
// PEAK (same), NOURISHED (dominant feeds officer), FEEDING (officer feeds dominant),
// SUPPRESSED (dominant controls officer), RESISTANT (officer controls dominant).
func DriftMultiplier(officerEssence, dominantElement string) float64 {
	o, d := officerEssence, dominantElement
	switch {
	case o == d:
		return 1.25 // PEAK
	case generation[d] == o:
		return 1.15 // NOURISHED
	case generation[o] == d:
		return 1.10 // FEEDING
	case control[d] == o:
		return 0.80 // SUPPRESSED
	case control[o] == d:
		return 0.90 // RESISTANT
	default:
		return 1.0
	}
}

// EffectiveStat applies the drift multiplier to a base stat, clamped to [1, 100].
func EffectiveStat(base int, multiplier float64) int {
	v := int(math.Round(float64(base) * multiplier))
	if v < 1 {
		return 1
	}
	if v > 100 {
		return 100
	}
	return v
}
