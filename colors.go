package mosaic

import (
	"image"
	"image/color"
	"math"

	color_extractor "github.com/marekm4/color-extractor"
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
		SmallBucket: 0.01,
		DownSizeTo:  224,
	})
	if len(colors) == 0 {
		return math.MaxUint32
	}
	return colorRGB(colors[0])
}
