package okcolor

import (
	"image/color"
	"math"
)

type LinearRGBA struct {
	R float64
	G float64
	B float64
	A uint16
}

var LinearRGBAModel = color.ModelFunc(linearRGBAConvert)

func linearRGBAConvert(c color.Color) color.Color {
	if _, ok := c.(LinearRGBA); ok {
		return c
	}

	return sRGBToLinearRGB(color.RGBA64Model.Convert(c).(color.RGBA64))
}

func (lc LinearRGBA) ClippedRGBA(clipFunc LinearRGBAClipper) (uint32, uint32, uint32, uint32) {
	return linearRGBToSRGB(clipFunc(lc)).RGBA()
}

func (lc LinearRGBA) RGBA() (uint32, uint32, uint32, uint32) {
	return lc.ClippedRGBA(LinearRGBAGamutClipperAdaptive05(0.05))
}

func linearRGBToSRGB(lc LinearRGBA) color.RGBA64 {
	return color.RGBA64{
		R: uint16(fromLinear(lc.R) * 65535),
		G: uint16(fromLinear(lc.G) * 65535),
		B: uint16(fromLinear(lc.B) * 65535),
		A: lc.A,
	}
}

func sRGBToLinearRGB(c color.RGBA64) LinearRGBA {
	return LinearRGBA{
		R: toLinear(float64(c.R) / 65535),
		G: toLinear(float64(c.G) / 65535),
		B: toLinear(float64(c.B) / 65535),
		A: c.A,
	}
}

func toLinear(x float64) float64 {
	if x >= 0.04045 {
		return math.Pow((x+0.055)/1.055, 2.4)
	} else {
		return x / 12.92
	}
}

const pow float64 = 1.0 / 2.4

func fromLinear(x float64) float64 {
	if x >= 0.0031308 {
		return math.Pow(x, pow)*1.055 - 0.055
	} else {
		return x * 12.92
	}
}

type LinearRGBAClipper func(LinearRGBA) LinearRGBA

func LinearRGBAGamutClipPreserveChroma(lc LinearRGBA) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipPreserveChroma(labConvert(lc).(Lab)).LinearRGBA(nil)
}

func LinearRGBAGamutClipProjectTo05(lc LinearRGBA) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipProjectTo05(labConvert(lc).(Lab)).LinearRGBA(nil)
}

func LinearRGBAGamutClipperProjectToL0(L0 float64) LinearRGBAClipper {
	return func(lc LinearRGBA) LinearRGBA {
		return LinearRGBAGamutClipProjectToL0(lc, L0)
	}
}

func LinearRGBAGamutClipProjectToL0(lc LinearRGBA, L0 float64) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipProjectToL0(labConvert(lc).(Lab), L0).LinearRGBA(nil)
}

func LinearRGBAGamutClipProjectToLCusp(lc LinearRGBA) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipProjectToLCusp(labConvert(lc).(Lab)).LinearRGBA(nil)
}

func LinearRGBAGamutClipperAdaptive05(alpha float64) LinearRGBAClipper {
	return func(lc LinearRGBA) LinearRGBA {
		return LinearRGBAGamutClipAdaptive05(lc, alpha)
	}
}

func LinearRGBAGamutClipAdaptive05(lc LinearRGBA, alpha float64) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipAdaptive05(labConvert(lc).(Lab), alpha).LinearRGBA(nil)
}

func LinearRGBAGamutClipperAdaptiveLCusp(alpha float64) LinearRGBAClipper {
	return func(lc LinearRGBA) LinearRGBA {
		return LinearRGBAGamutClipAdaptiveLCusp(lc, alpha)
	}
}

func LinearRGBAGamutClipAdaptiveLCusp(lc LinearRGBA, alpha float64) LinearRGBA {
	if (lc.R >= 0) && (lc.R <= 1) && (lc.G >= 0) && (lc.G <= 1) && (lc.B >= 0) && (lc.B <= 1) {
		return lc
	}
	return GamutClipAdaptiveLCusp(labConvert(lc).(Lab), alpha).LinearRGBA(nil)
}
