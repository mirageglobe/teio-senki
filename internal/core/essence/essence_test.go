package essence_test

import (
	"testing"

	"github.com/mirageglobe/teio-senki/internal/core/essence"
)

func TestDriftMultiplier(t *testing.T) {
	cases := []struct {
		officer, dominant string
		want              float64
	}{
		{"Fire", "Fire", 1.25},   // PEAK
		{"Fire", "Wood", 1.15},   // NOURISHED: Wood generates Fire
		{"Wood", "Fire", 1.10},   // FEEDING: Wood feeds Fire
		{"Wood", "Metal", 0.80},  // SUPPRESSED: Metal controls Wood
		{"Metal", "Wood", 0.90},  // RESISTANT: Metal controls Wood (reversed)
		{"Water", "Metal", 1.15}, // NOURISHED: Metal generates Water
	}
	for _, c := range cases {
		got := essence.DriftMultiplier(c.officer, c.dominant)
		if got != c.want {
			t.Errorf("officer=%s dominant=%s: want %.2f, got %.2f", c.officer, c.dominant, c.want, got)
		}
	}
}

func TestEffectiveStat(t *testing.T) {
	if got := essence.EffectiveStat(80, 1.25); got != 100 {
		t.Errorf("clamp to 100: got %d", got)
	}
	if got := essence.EffectiveStat(100, 0.80); got != 80 {
		t.Errorf("expected 80: got %d", got)
	}
	if got := essence.EffectiveStat(1, 0.10); got != 1 {
		t.Errorf("clamp to 1: got %d", got)
	}
}
