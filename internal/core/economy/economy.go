// Package economy owns city yield and corruption formulas.
// It does NOT own city state or resource tracking.
package economy

const techBonusMax = 0.5

// GrainYield calculates agriculture output boosted by technology.
func GrainYield(agriculture, technology int) int {
	bonus := float64(technology) / 100.0 * techBonusMax
	v := int(float64(agriculture) * (1.0 + bonus))
	if v < 0 {
		return 0
	}
	return v
}

// GoldYield calculates commerce output boosted by technology.
func GoldYield(commerce, technology int) int {
	bonus := float64(technology) / 100.0 * techBonusMax
	v := int(float64(commerce) * (1.0 + bonus))
	if v < 0 {
		return 0
	}
	return v
}

// CorruptionLoss is a flat drain on yields. order=100 → 0 loss; order=0 → 100 loss.
func CorruptionLoss(order int) int {
	if order > 100 {
		order = 100
	}
	if order < 0 {
		order = 0
	}
	return 100 - order
}
