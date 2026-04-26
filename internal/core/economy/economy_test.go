package economy_test

import (
	"testing"

	"github.com/mirageglobe/teio-senki/internal/core/economy"
)

func TestGrainYield(t *testing.T) {
	// no technology: 50 * 1.0 = 50
	if got := economy.GrainYield(50, 0); got != 50 {
		t.Errorf("want 50, got %d", got)
	}
	// max tech: 50 * 1.5 = 75
	if got := economy.GrainYield(50, 100); got != 75 {
		t.Errorf("want 75, got %d", got)
	}
	// zero agriculture always yields 0
	if got := economy.GrainYield(0, 100); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestGoldYield(t *testing.T) {
	if got := economy.GoldYield(40, 0); got != 40 {
		t.Errorf("want 40, got %d", got)
	}
	if got := economy.GoldYield(40, 100); got != 60 {
		t.Errorf("want 60, got %d", got)
	}
}

func TestCorruptionLoss(t *testing.T) {
	if got := economy.CorruptionLoss(100); got != 0 {
		t.Errorf("want 0 loss at max order, got %d", got)
	}
	if got := economy.CorruptionLoss(0); got != 100 {
		t.Errorf("want 100 loss at zero order, got %d", got)
	}
	if got := economy.CorruptionLoss(60); got != 40 {
		t.Errorf("want 40, got %d", got)
	}
}
