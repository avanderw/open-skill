// Package rating implements the Glicko-2 rating system, ported from the
// canonical open-source implementation (glicko2 Python package) and matching
// the Glickman paper: http://www.glicko.net/glicko/glicko2.pdf
//
// A rated competitor in a mode has three values:
//   - rating  (mu)  — displayed rating.
//   - rd      (phi) — rating deviation; uncertainty (large for new players).
//   - vol     (sigma) — volatility; how much the rating can change. Not shown.
//
// q = ln(10)/400 is folded into the scale factor SCALE = 400/ln(10) = 173.7178
// so that g() has no separate q term, reproducing the published worked example
// (rating 1464.05, RD 151.52).
package rating

import "math"

const (
	DefaultRating = 1500.0   // new-player rating anchor
	DefaultRd     = 350.0    // new-player rating deviation
	DefaultVol    = 0.06     // new-player volatility
	DefaultTau    = 0.5      // constraint on volatility change
	scale         = 173.7178 // 400 / ln(10)
)

// Engine is a configured Glicko-2 calculator.
type Engine struct{ Tau float64 }

// New returns a default engine.
func New() Engine { return Engine{Tau: DefaultTau} }

// Player is a competitor's rating state in one mode.
type Player struct {
	Mu    float64 // rating
	RD    float64 // rating deviation
	Sigma float64 // volatility
}

// NewPlayer returns the rating state of a brand-new competitor.
func NewPlayer() Player {
	return Player{Mu: DefaultRating, RD: DefaultRd, Sigma: DefaultVol}
}

// Opponent is a rated opponent a player faced in a rating period.
type Opponent struct {
	Mu    float64
	RD    float64
	Score float64
}

// Result is a player's updated state after a rating period.
type Result struct {
	Mu    float64
	RD    float64
	Sigma float64
}

// g is the Glicko-2 g(RD) function (reference _g; no q term — folded into scale).
func g(rd float64) float64 {
	return 1 / math.Sqrt(1+3*rd*rd/(math.Pi*math.Pi))
}

// eVal is the Glicko-2 expectation E (reference _E), all on the scaled scale.
func eVal(pRating, opRating, opRD float64) float64 {
	return 1 / (1 + math.Exp(-1*g(opRD)*(pRating-opRating)))
}

// Update computes the new rating state for player r given opponents o in one
// rating period. Port mirrors the reference implementation exactly.
func (e Engine) Update(r Player, o []Opponent) Result {
	// Convert the player and opponents to the internal Glicko-2 scale.
	pRat := (r.Mu - DefaultRating) / scale
	pRD := r.RD / scale
	pVol := r.Sigma

	var rat []float64
	var rdL []float64
	var scores []float64
	for _, op := range o {
		rat = append(rat, (op.Mu-DefaultRating)/scale)
		rdL = append(rdL, op.RD/scale)
		scores = append(scores, op.Score)
	}

	// v = 1 / sum(g^2 * E * (1-E))
	var vSum float64
	for i := range rat {
		e := eVal(pRat, rat[i], rdL[i])
		gv := g(rdL[i])
		vSum += gv * gv * e * (1 - e)
	}
	v := 1.0 / math.Max(vSum, 1e-12)

	// delta = v * sum(g * (score - E))
	var dSum float64
	for i := range rat {
		dSum += g(rdL[i]) * (scores[i] - eVal(pRat, rat[i], rdL[i]))
	}
	delta := v * dSum

	// --- new volatility (sigma') via the reference's iterative solve ---
	a := math.Log(pVol * pVol)
	phiSq := pRD * pRD

	f := func(x float64) float64 {
		ex := math.Exp(x)
		return ex*(delta*delta-phiSq-v-ex)/(2*math.Pow(phiSq+v+ex, 2)) - (x-a)/(e.Tau*e.Tau)
	}

	eps := 1e-6
	A := a
	var B float64
	if delta*delta > phiSq+v {
		B = math.Log(delta*delta - phiSq - v)
	} else {
		k := 1.0
		for f(a-k*math.Sqrt(e.Tau*e.Tau)) < 0 {
			k++
		}
		B = a - k*math.Sqrt(e.Tau*e.Tau)
	}
	fA := f(A)
	fB := f(B)
	for math.Abs(B-A) > eps {
		C := A + (A-B)*fA/(fB-fA)
		fC := f(C)
		if fC*fB <= 0 {
			A = B
			fA = fB
		} else {
			fA = fA / 2.0
		}
		B = C
		fB = fC
	}
	sigmaPrime := math.Exp(A / 2)

	// --- new RD ---
	newRD := math.Sqrt(pRD*pRD + sigmaPrime*sigmaPrime)
	newRD = 1 / math.Sqrt(1/(newRD*newRD)+1/v)

	// --- new rating ---
	var rSum float64
	for i := range rat {
		rSum += g(rdL[i]) * (scores[i] - eVal(pRat, rat[i], rdL[i]))
	}
	newRat := pRat + newRD*newRD*rSum

	return Result{
		Mu:    DefaultRating + scale*newRat,
		RD:    scale * newRD,
		Sigma: sigmaPrime,
	}
}
