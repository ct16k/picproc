// based on:
// https://bottosson.github.io/posts/gamutclipping/

package okcolor

import "math"

const eps = 0.00001

type Clipper func(Lab) Lab

func GamutClipPreserveChroma(lc Lab) Lab {
	return GamutClipProjectToL0(lc, clamp(lc.L, 0, 1))
}

func GamutClipProjectTo05(lc Lab) Lab {
	return GamutClipProjectToL0(lc, 0.5)
}

func GamutClipperProjectToL0(L0 float64) Clipper {
	return func(lc Lab) Lab {
		return GamutClipProjectToL0(lc, L0)
	}
}

func GamutClipProjectToL0(lc Lab, L0 float64) Lab {
	c := max(eps, math.Sqrt(lc.A*lc.A+lc.B*lc.B))
	a_ := lc.A / c
	b_ := lc.B / c

	t := findGamutIntersection(a_, b_, lc.L, c, L0)
	lClipped := L0*(1-t) + t*lc.L
	cClipped := t * c

	return Lab{
		L:     lClipped,
		A:     cClipped * a_,
		B:     cClipped * b_,
		Alpha: lc.Alpha,
	}
}

func GamutClipProjectToLCusp(lc Lab) Lab {
	c := max(eps, math.Sqrt(lc.A*lc.A+lc.B*lc.B))
	a_ := lc.A / c
	b_ := lc.B / c

	lC, cC := findCusp(a_, b_)

	l0 := lC

	t := findGamutIntersectionWithCusp(a_, b_, lc.L, c, l0, lC, cC)

	lClipped := l0*(1-t) + t*lc.L
	cClipped := t * c

	return Lab{
		L:     lClipped,
		A:     cClipped * a_,
		B:     cClipped * b_,
		Alpha: lc.Alpha,
	}
}

func GamutClipperAdaptive05(alpha float64) Clipper {
	return func(lc Lab) Lab {
		return GamutClipAdaptive05(lc, alpha)
	}
}

func GamutClipAdaptive05(lc Lab, alpha float64) Lab {
	c := max(eps, math.Sqrt(lc.A*lc.A+lc.B*lc.B))
	a_ := lc.A / c
	b_ := lc.B / c

	ld := lc.L - 0.5
	e1 := 0.5 + math.Abs(ld) + alpha*c
	L0 := 0.5 * (1 + sgn(ld)*(e1-math.Sqrt(e1*e1-2*math.Abs(ld))))

	t := findGamutIntersection(a_, b_, lc.L, c, L0)
	lClipped := L0*(1-t) + t*lc.L
	cClipped := t * c

	return Lab{
		L:     lClipped,
		A:     cClipped * a_,
		B:     cClipped * b_,
		Alpha: lc.Alpha,
	}
}

func GamutClipperAdaptiveLCusp(alpha float64) Clipper {
	return func(lc Lab) Lab {
		return GamutClipAdaptiveLCusp(lc, alpha)
	}
}

func GamutClipAdaptiveLCusp(lc Lab, alpha float64) Lab {
	c := max(eps, math.Sqrt(lc.A*lc.A+lc.B*lc.B))
	a_ := lc.A / c
	b_ := lc.B / c

	lC, cC := findCusp(a_, b_)

	ld := lc.L - lC
	var k float64
	if ld > 0 {
		k = 2 * (1 - lC)
	} else {
		k = 2 * lC
	}

	e1 := 0.5*k + math.Abs(ld) + alpha*c/k
	l0 := lC + 0.5*(sgn(ld)*(e1-math.Sqrt(e1*e1-2*k*math.Abs(ld))))

	t := findGamutIntersectionWithCusp(a_, b_, lc.L, c, l0, lC, cC)
	lClipped := l0*(1-t) + t*lc.L
	cClipped := t * c

	return Lab{
		L:     lClipped,
		A:     cClipped * a_,
		B:     cClipped * b_,
		Alpha: lc.Alpha,
	}
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	} else if x > max {
		return max
	} else {
		return x
	}
}

func sgn(x float64) float64 {
	if x < 0 {
		return -1
	} else if x > 1 {
		return 1
	}
	return 0
}

// findGamutIntersection finds intersection of the line defined by
// L = L0 * (1 - t) + t * L1
// C = t * C1
// a and b must be normalized so a^2 + b^2 == 1
func findGamutIntersection(a, b, L1, C1, L0 float64) float64 {
	// find the cusp of the gamut triangle
	lC, cC := findCusp(a, b)
	return findGamutIntersectionWithCusp(a, b, L1, C1, L0, lC, cC)
}

func findGamutIntersectionWithCusp(a, b, L1, C1, L0, lC, cC float64) float64 {
	// find the intersection for upper and lower half separately
	var t float64
	if ((L1-L0)*cC - (lC-L0)*C1) <= 0 { // lower half

		t = cC * L0 / (C1*lC + cC*(L0-L1))
	} else { // upper half
		// first intersect with triangle
		t = cC * (L0 - 1) / (C1*(lC-1) + cC*(L0-L1))

		// then one step Halley's method
		dL := L1 - L0
		dC := C1

		kL := +0.3963377774*a + 0.2158037573*b
		kM := -0.1055613458*a - 0.0638541728*b
		kS := -0.0894841775*a - 1.2914855480*b

		lDt := dL + dC*kL
		mDt := dL + dC*kM
		sDt := dL + dC*kS

		// if higher accuracy is required, 2 or 3 iterations of the following block can be used
		L := L0*(1-t) + t*L1
		C := t * C1

		l_ := L + C*kL
		l := l_ * l_ * l_
		ldt := 3 * lDt * l_ * l_
		ldt2 := 6 * lDt * lDt * l_

		m_ := L + C*kM
		m := m_ * m_ * m_
		mdt := 3 * mDt * m_ * m_
		mdt2 := 6 * mDt * mDt * m_

		s_ := L + C*kS
		s := s_ * s_ * s_
		sdt := 3 * sDt * s_ * s_
		sdt2 := 6 * sDt * sDt * s_

		r := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s - 1
		r1 := 4.0767416621*ldt - 3.3077115913*mdt + 0.2309699292*sdt
		r2 := 4.0767416621*ldt2 - 3.3077115913*mdt2 + 0.2309699292*sdt2

		uR := r1 / (r1*r1 - 0.5*r*r2)
		tR := math.MaxFloat64
		if uR >= 0 {
			tR = -r * uR
		}

		g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s - 1
		g1 := -1.2684380046*ldt + 2.6097574011*mdt - 0.3413193965*sdt
		g2 := -1.2684380046*ldt2 + 2.6097574011*mdt2 - 0.3413193965*sdt2

		uG := g1 / (g1*g1 - 0.5*g*g2)
		tG := math.MaxFloat64
		if uG >= 0 {
			tG = -g * uG
		}

		b := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s - 1
		b1 := -0.0041960863*ldt - 0.7034186147*mdt + 1.7076147010*sdt
		b2 := -0.0041960863*ldt2 - 0.7034186147*mdt2 + 1.7076147010*sdt2

		uB := b1 / (b1*b1 - 0.5*b*b2)
		tB := math.MaxFloat64
		if uB >= 0 {
			tB = -b * uB
		}

		t += min(tR, tG, tB)
		// end
	}

	return t
}

// findCusp finds lCusp and cCusp for a given hue
// a and b must be normalized so a^2 + b^2 == 1
func findCusp(a, b float64) (float64, float64) {
	// first, find the maximum saturation (saturation S = C/L)
	sCusp := computeMaxSaturation(a, b)

	// convert to linear sRGB to find the first point where at least one of r,g or b >= 1:
	rgbAtMax := Lab{
		L: 1,
		A: sCusp * a,
		B: sCusp * b,
	}.LinearRGBA(nil)
	lCusp := math.Cbrt(1 / max(max(rgbAtMax.R, rgbAtMax.G), rgbAtMax.B))
	cCusp := lCusp * sCusp

	return lCusp, cCusp
}

// computeMaxSaturation finds the maximum saturation possible for a given hue that fits in sRGB
// Saturation here is defined as S = C/L
// a and b must be normalized so a^2 + b^2 == 1
func computeMaxSaturation(a, b float64) float64 {
	// max saturation will be when one of r, g or b goes below zero.

	// select different coefficients depending on which component goes below zero first
	var k0, k1, k2, k3, k4, wl, wm, ws float64
	if (-1.88170328*a - 0.80936493*b) > 1 { // red component
		k0 = +1.19086277
		k1 = +1.76576728
		k2 = +0.59662641
		k3 = +0.75515197
		k4 = +0.56771245

		wl = +4.0767416621
		wm = -3.3077115913
		ws = +0.2309699292
	} else if (1.81444104*a - 1.19445276*b) > 1 { // green component
		k0 = +0.73956515
		k1 = -0.45954404
		k2 = +0.08285427
		k3 = +0.12541070
		k4 = +0.14503204

		wl = -1.2684380046
		wm = +2.6097574011
		ws = -0.3413193965
	} else { // blue component
		k0 = +1.35733652
		k1 = -0.00915799
		k2 = -1.15130210
		k3 = -0.50559606
		k4 = +0.00692167

		wl = -0.0041960863
		wm = -0.7034186147
		ws = +1.7076147010
	}

	// approximate max saturation using a polynomial:
	sat := k0 + k1*a + k2*b + k3*a*a + k4*a*b

	// do one step Halley's method to get closer
	// this gives an error less than 10e6, except for some blue hues where the dS/dh is close to infinite
	// this should be sufficient for most applications, otherwise do two/three steps

	kL := +0.3963377774*a + 0.2158037573*b
	kM := -0.1055613458*a - 0.0638541728*b
	kS := -0.0894841775*a - 1.2914855480*b

	// start
	l_ := 1 + sat*kL
	m_ := 1 + sat*kM
	s_ := 1 + sat*kS

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	lDS := 3 * kL * l_ * l_
	mDS := 3 * kM * m_ * m_
	sDS := 3 * kS * s_ * s_

	lDS2 := 6 * kL * kL * l_
	mDS2 := 6 * kM * kM * m_
	sDS2 := 6 * kS * kS * s_

	f := wl*l + wm*m + ws*s
	f1 := wl*lDS + wm*mDS + ws*sDS
	f2 := wl*lDS2 + wm*mDS2 + ws*sDS2

	sat = sat - f*f1/(f1*f1-0.5*f*f2)
	// end

	return sat
}
