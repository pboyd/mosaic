package mosaic

import (
	"image"
	"image/color"
	"math"

	//color_extractor "github.com/marekm4/color-extractor"
	color_extractor "github.com/pboyd/color-extractor"
)

func colorRGB(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	return (r << 16) | (g << 8) | b
}

func colorVector(c uint32) []float64 {
	return []float64{
		float64(c >> 16),         // r
		float64((c >> 8) & 0xff), // g
		float64(c & 0xff),        // b
	}
}

func vectorColor(v []float64) uint32 {
	return uint32(v[0])<<16 | uint32(v[1])<<8 | uint32(v[2])
}

func primaryColor(img image.Image) uint32 {
	colors := color_extractor.ExtractColorsWithConfig(img, color_extractor.Config{
		DownSizeTo:  100,
		SmallBucket: 0,
	})
	return colorRGB(colors[0])
}

func averageColor(img image.Image) uint32 {
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

	r >>= 8
	g >>= 8
	b >>= 8

	return (r << 16) | (g << 8) | b
}

// colorDistance returns the squared distance between two colors.
func colorDistance(a, b uint32) uint32 {
	// RGB is kind of like a 3D vector, so we can use the distance formula.

	rA, gA, bA := uint8(a>>16), uint8(a>>8), uint8(a)
	rB, gB, bB := uint8(b>>16), uint8(b>>8), uint8(b)

	var rd, gd, bd uint8
	if rA > rB {
		rd = rA - rB
	} else {
		rd = rB - rA
	}
	if gA > gB {
		gd = gA - gB
	} else {
		gd = gB - gA
	}
	if bA > bB {
		bd = bA - bB
	} else {
		bd = bB - bA
	}

	// Squared distance is good enough for our purposes.
	return uint32(rd)*uint32(rd) + uint32(gd)*uint32(gd) + uint32(bd)*uint32(bd)
}
