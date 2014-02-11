package classifier

import (
	"math"
	"testing"
)

func almostEqual(a, b, d float64) bool {
	return math.Abs(a-b) < d
}

func TestInvChi2(t *testing.T) {
	c2 := invChi2(5, 4)
	if !almostEqual(c2, 0.2873, 0.0001) {
		t.Errorf("Chi2(5, 4) returned %f instead of 0.2873+=0.0001", c2)
	}
}
