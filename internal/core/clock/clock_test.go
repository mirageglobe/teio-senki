package clock_test

import (
	"testing"

	"github.com/mirageglobe/teio-senki/internal/core/clock"
)

func TestAdvanceYearRollover(t *testing.T) {
	s := clock.NewAt(189, 12)
	next := clock.Advance(s)
	if next.Year != 190 || next.Month != 1 {
		t.Errorf("want 190.1, got %d.%d", next.Year, next.Month)
	}
	if next.TotalMonths != s.TotalMonths+1 {
		t.Errorf("TotalMonths not incremented")
	}
}

func TestElementForMonth(t *testing.T) {
	cases := []struct {
		month int
		want  string
	}{
		{1, "Water"},  // Zi
		{3, "Wood"},   // Yin
		{6, "Fire"},   // Si
		{9, "Metal"},  // Shen
		{11, "Earth"}, // Xu
	}
	for _, c := range cases {
		if got := clock.ElementForMonth(c.month); got != c.want {
			t.Errorf("month %d: want %s, got %s", c.month, c.want, got)
		}
	}
}

func TestStemForYear(t *testing.T) {
	if got := clock.StemForYear(184); got != "Jia" {
		t.Errorf("184: want Jia, got %s", got)
	}
	// (189-184)%10 = 5 → "Ji"
	if got := clock.StemForYear(189); got != "Ji" {
		t.Errorf("189: want Ji, got %s", got)
	}
}
