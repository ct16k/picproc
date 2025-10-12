package palette

import (
	"fmt"
	"image/color"
	"io"
	"math"

	"picproc/okcolor"
)

type Lab []okcolor.Lab

var (
	_ PaletteRIFFReaderWriter = &Lab{}
	_ PaletteConverter        = &Lab{}
)

func NewLabPalette(p color.Palette) *Lab {
	pal := &Lab{}
	pal.From(p)
	return nil
}

func (p *Lab) Convert(lc okcolor.Lab) okcolor.Lab {
	if len(*p) == 0 {
		return okcolor.Lab{}
	}
	return (*p)[p.Index(lc)]
}

func (p *Lab) Index(lc okcolor.Lab) int {
	ret, bestSum := 0, math.MaxFloat64
	for i, v := range *p {
		dL := lc.L - v.L
		da := lc.A - v.A
		db := lc.B - v.B
		dA := lc.Alpha - v.Alpha
		sum := dL*dL + da*da + db*db + float64(dA*dA)
		if sum < bestSum {
			if sum == 0 {
				return i
			}
			ret, bestSum = i, sum
		}
	}
	return ret
}

func (p *Lab) From(pal color.Palette) int64 {
	for _, col := range pal {
		*p = append(*p, okcolor.LabModel.Convert(col).(okcolor.Lab))
	}

	return int64(len(pal))
}

func (p *Lab) ReadRIFF(r io.Reader) (int64, error) {
	pals, err := ReadFrom(r)
	if err != nil {
		return 0, fmt.Errorf("could not load palettes: %w", err)
	}

	var n int64
	for _, pal := range pals {
		n += p.From(pal)
	}

	return n, nil
}

func (p *Lab) To(m color.Model) (int64, color.Palette) {
	var pal color.Palette
	n := p.ToPalette(m, &pal)

	return int64(n), pal
}

func (p *Lab) ToPalette(m color.Model, pal *color.Palette) int64 {
	n, pc, pn := len(*p), cap(*pal), len(*pal)
	if pc < n+pn {
		*pal = append(make(color.Palette, 0, n+pn), (*pal)...)
	}

	for _, lc := range *p {
		(*pal)[n] = m.Convert(lc)
		n++
	}

	return int64(pn)
}

func (p *Lab) WriteRIFF(w io.Writer) (int64, error) {
	pal := make([]color.Color, len(*p))
	for i, lc := range *p {
		pal[i] = color.RGBAModel.Convert(lc)
	}

	if n, err := WriteTo(w, []color.Palette{pal}); err != nil {
		return n, fmt.Errorf("could not save palette: %w", err)
	} else {
		return n, nil
	}
}

type LinearRGBA []okcolor.LinearRGBA

var (
	_ PaletteRIFFReaderWriter = &LinearRGBA{}
	_ PaletteConverter        = &LinearRGBA{}
)

func NewLinearRGBAPalette(p color.Palette) *LinearRGBA {
	pal := &LinearRGBA{}
	pal.From(p)
	return nil
}

func (p *LinearRGBA) Convert(lc okcolor.LinearRGBA) okcolor.LinearRGBA {
	if len(*p) == 0 {
		return okcolor.LinearRGBA{}
	}
	return (*p)[p.Index(lc)]
}

func (p *LinearRGBA) Index(lc okcolor.LinearRGBA) int {
	ret, bestSum := 0, math.MaxFloat64
	for i, v := range *p {
		dr := lc.R - v.R
		dg := lc.G - v.G
		db := lc.B - v.B
		dA := lc.A - v.A
		sum := dr*dr + dg*dg + db*db + float64(dA*dA)
		if sum < bestSum {
			if sum == 0 {
				return i
			}
			ret, bestSum = i, sum
		}
	}
	return ret
}

func (p *LinearRGBA) From(pal color.Palette) int64 {
	for _, col := range pal {
		*p = append(*p, okcolor.LinearRGBAModel.Convert(col).(okcolor.LinearRGBA))
	}

	return int64(len(pal))
}

func (p *LinearRGBA) ReadRIFF(r io.Reader) (int64, error) {
	pals, err := ReadFrom(r)
	if err != nil {
		return 0, fmt.Errorf("could not load palettes: %w", err)
	}

	n := 0
	for _, pal := range pals {
		n += len(pal)
		for _, col := range pal {
			*p = append(*p, okcolor.LinearRGBAModel.Convert(col).(okcolor.LinearRGBA))
		}
	}

	return int64(n), nil
}

func (p *LinearRGBA) To(m color.Model) (int64, color.Palette) {
	var pal color.Palette
	n := p.ToPalette(m, &pal)

	return int64(n), pal
}

func (p *LinearRGBA) ToPalette(m color.Model, pal *color.Palette) int64 {
	n, pc, pn := len(*p), cap(*pal), len(*pal)
	if pc < n+pn {
		*pal = append(make(color.Palette, 0, n+pn), (*pal)...)
	}

	for _, lc := range *p {
		(*pal)[n] = m.Convert(lc)
		n++
	}

	return int64(pn)
}

func (p *LinearRGBA) WriteRIFF(w io.Writer) (int64, error) {
	pal := make([]color.Color, len(*p))
	for i, lc := range *p {
		pal[i] = color.RGBAModel.Convert(lc)
	}

	if n, err := WriteTo(w, []color.Palette{pal}); err != nil {
		return n, fmt.Errorf("could not save palette: %w", err)
	} else {
		return n, nil
	}
}
