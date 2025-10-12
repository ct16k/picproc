package mangle

import (
	"image"
	"log/slog"

	"picproc/palette"

	"golang.org/x/image/draw"
)

func repallete(logger *slog.Logger, img image.Image, palName string, dither bool) (image.Image, error) {
	pal, err := palette.LoadPalette(palName)
	if err != nil {
		return nil, err
	}

	logger.Info("applying palette", "colors", len(pal))
	sr := img.Bounds()
	dr := image.Rect(0, 0, sr.Dx(), sr.Dy())
	dest := image.NewPaletted(dr, pal)

	if dither {
		draw.FloydSteinberg.Draw(dest, dr, img, sr.Min)
	} else {
		draw.Draw(dest, dr, img, dr.Min, draw.Src)
	}
	return dest, err
}
