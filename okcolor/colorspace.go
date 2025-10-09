// based on:
// https://bottosson.github.io/posts/oklab/
// https://bottosson.github.io/posts/colorwrong/#what-can-we-do%3F

package okcolor

import (
	"image/color"
	"math"
)

type Lab struct {
	L     float64 // perceived lightness
	A     float64 // how green/red the color is
	B     float64 // how blue/yellow the color is
	Alpha uint16  // alpha
}

var LabModel = color.ModelFunc(labConvert)

func labConvert(c color.Color) color.Color {
	switch lc := c.(type) {
	case Lab:
		return c
	case LCh:
		return lc.Lab()
	}

	col := linearRGBAConvert(c).(LinearRGBA)

	var l, m, s float64
	l = math.Cbrt(0.4122214708*col.R + 0.5363325363*col.G + 0.0514459929*col.B)
	m = math.Cbrt(0.2119034982*col.R + 0.6806995451*col.G + 0.1073969566*col.B)
	s = math.Cbrt(0.0883024619*col.R + 0.2817188376*col.G + 0.6299787005*col.B)

	return Lab{
		L:     0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		A:     1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		B:     0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
		Alpha: col.A,
	}
}

func (lc Lab) RGBA() (uint32, uint32, uint32, uint32) {
	c := linearRGBToSRGB(lc.LinearRGBA(GamutClipperAdaptive05(0.05)))
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}

func (lc Lab) LinearRGBA(clipFunc Clipper) LinearRGBA {
	var l, m, s float64
	l = lc.L + 0.3963377774*lc.A + 0.2158037573*lc.B
	l = l * l * l
	m = lc.L - 0.1055613458*lc.A - 0.0638541728*lc.B
	m = m * m * m
	s = lc.L - 0.0894841775*lc.A - 1.2914855480*lc.B
	s = s * s * s

	r := +4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	b := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	if (clipFunc != nil) && ((r < 0) || (r > 1) || (g < 0) || (g > 1) || (b < 0) || (b > 1)) {
		return clipFunc(lc).LinearRGBA(nil)
	}

	return LinearRGBA{
		R: r,
		G: g,
		B: b,
		A: lc.Alpha,
	}
}

func (lc Lab) LCh() LCh {
	return LCh{
		L:     lc.L,
		C:     math.Sqrt((lc.A * lc.A) + (lc.B * lc.B)),
		H:     math.Atan2(lc.B, lc.A),
		Alpha: lc.Alpha,
	}
}

type LCh struct {
	L     float64 // perceived lightness
	C     float64 // chroma
	H     float64 // hue
	Alpha uint16  // alpha
}

var LChModel = color.ModelFunc(lchConvert)

func lchConvert(c color.Color) color.Color {
	switch lc := c.(type) {
	case *LCh:
		return c
	case Lab:
		return lc.LCh()
	}

	return labConvert(c).(Lab).LCh()
}

func (lc LCh) RGBA() (uint32, uint32, uint32, uint32) {
	return lc.Lab().RGBA()
}

func (lc LCh) LinearRGBA(clipFunc Clipper) LinearRGBA {
	return lc.Lab().LinearRGBA(clipFunc)
}

func (lc LCh) Lab() Lab {
	return Lab{
		L:     lc.L,
		A:     lc.C * math.Cos(lc.H),
		B:     lc.C * math.Sin(lc.H),
		Alpha: lc.Alpha,
	}
}
