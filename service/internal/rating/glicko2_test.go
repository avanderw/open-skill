package rating

import (
	"math"
	"testing"
)

// TestGlicko2Example reproduces the Glickman paper's worked example
// (http://www.glicko.net/glicko/glicko2.pdf, "Example of the Glicko-2 system").
//
// Player: rating 1500, RD 200, volatility 0.06, tau 0.5.
// Opponents (one rating period):
//
//	(1400, RD 30,   win), (1550, RD 100, loss), (1700, RD 300, loss)
//
// Expected: rating 1464.05, RD 151.52, volatility ~0.06.
func TestGlicko2Example(t *testing.T) {
	res := (Engine{Tau: 0.5}).Update(
		Player{Mu: 1500, RD: 200, Sigma: 0.06},
		[]Opponent{
			{Mu: 1400, RD: 30, Score: 1.0},
			{Mu: 1550, RD: 100, Score: 0.0},
			{Mu: 1700, RD: 300, Score: 0.0},
		},
	)
	if got, want := res.Mu, 1464.05; math.Abs(got-want) > 0.5 {
		t.Errorf("rating = %v, want %v", got, want)
	}
	if got, want := res.RD, 151.52; math.Abs(got-want) > 0.5 {
		t.Errorf("rd = %v, want %v", got, want)
	}
	if got, want := res.Sigma, 0.06; math.Abs(got-want) > 0.01 {
		t.Errorf("vol = %v, want ~%v", got, want)
	}
}

// TestSingleGameSwing documents that Glicko-2 produces a large swing on a
// new player's first decisive game (both at max RD=350), but it must be bounded
// and the uncertainty must shrink.
func TestSingleGameSwing(t *testing.T) {
	res := New().Update(
		NewPlayer(),
		[]Opponent{{Mu: DefaultRating, RD: DefaultRd, Score: 1.0}},
	)
	if res.Mu > 1750 || res.Mu < 1550 {
		t.Errorf("winner after one game vs max-RD = %v, expect a bounded bump to ~1600-1700", res.Mu)
	}
	if res.RD >= DefaultRd {
		t.Errorf("winner RD should shrink after a game; got %v", res.RD)
	}
}
