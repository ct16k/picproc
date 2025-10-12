package okcolor

import (
	"image"
	"image/color"
)

type Image struct {
	// Pix holds the image's pixels, as palette indices. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*1].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
	// Palette is the image's palette.
	Palette color.Palette
}

// bytes per pixel: r, g, b, a float64 = 4 * 8 = 32

func NewImage(r image.Rectangle) *Image {
	return &Image{
		Pix:    make([]uint8, r.Dx()*r.Dy()*32),
		Stride: 32 * r.Dx(),
		Rect:   r,
	}
}
