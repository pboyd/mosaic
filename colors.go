package mosaic

import (
	"image"
	"image/color"
	"math"
)

func averageColor(img image.Image) color.Color {
	bounds := img.Bounds()

	count := bounds.Dx() * bounds.Dy()
	if count > math.MaxUint32 {
		panic("image too large")
	}

	var r, g, b uint32
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r += r1
			g += g1
			b += b1
		}
	}

	r /= uint32(count)
	g /= uint32(count)
	b /= uint32(count)

	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
}
