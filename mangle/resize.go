package mangle

import (
	"image"
	"image/color"
	"log/slog"
	"math"

	"golang.org/x/image/draw"
)

func resize(logger *slog.Logger, img image.Image, width, height int, crop bool, fillColor color.Color) (image.Image, error) {
	srcBounds := img.Bounds()
	srcWidth := float64(srcBounds.Dx())
	srcHeight := float64(srcBounds.Dy())

	destWidth := float64(width)
	if destWidth == 0 {
		destWidth = srcWidth
	}

	destHeight := float64(height)
	if destHeight == 0 {
		destHeight = srcHeight
	}

	if (srcWidth == destWidth) && (srcHeight == destHeight) {
		return img, nil
	}

	destSize := image.Rect(0, 0, int(destWidth), int(destHeight))
	destBounds := image.Rect(0, 0, int(destWidth), int(destHeight))

	srcAR := srcWidth / srcHeight
	destAR := destWidth / destHeight
	var fill bool
	if crop {
		if srcAR < destAR {
			dh := int(math.Round((srcHeight - srcWidth/destAR) / 2))
			srcBounds.Min.Y += dh
			srcBounds.Max.Y -= dh
		} else if srcAR > destAR {
			dw := int(math.Round((srcWidth - srcHeight*destAR) / 2))
			srcBounds.Min.X += dw
			srcBounds.Max.X -= dw
		}
	} else {
		if srcAR < destAR {
			dw := destHeight * srcAR
			if fillColor == nil {
				destSize.Max.X = int(math.Round(dw))
				destBounds.Max.X = destSize.Max.X
			} else {
				if fill = destWidth > dw; fill {
					idw := int(math.Round((destWidth - dw) / 2))
					destBounds.Min.X += idw
					destBounds.Max.X -= idw
				}
			}
		} else if srcAR > destAR {
			dh := destWidth / srcAR
			if fillColor == nil {
				destSize.Max.Y = int(math.Round(dh))
				destBounds.Max.Y = destSize.Max.Y
			} else {
				if fill = destHeight > dh; fill {
					idh := int(math.Round((destHeight - dh) / 2))
					destBounds.Min.Y += idh
					destBounds.Max.Y -= idh
				}
			}
		}
	}

	logger.Info("resizing", "width", destBounds.Dx(), "height", destBounds.Dy())
	dest := image.NewRGBA64(destSize)
	if fill && (fillColor != nil) {
		draw.Draw(dest, destSize, image.NewUniform(fillColor), destSize.Min, draw.Over)
	}
	draw.CatmullRom.Scale(dest, destBounds, img, srcBounds, draw.Over, nil)

	return dest, nil
}
