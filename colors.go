package mosaic

import (
	"errors"
	"image"
	"image/color"

	color_extractor "github.com/marekm4/color-extractor"
)

// ErrNoPrimaryColor is returned when no primary color is found in the image.
var ErrNoPrimaryColor = errors.New("no primary color found")

func colorVector(c uint32) []float64 {
	return []float64{
		float64(c >> 16),         // r
		float64((c >> 8) & 0xff), // g
		float64(c & 0xff),        // b
	}
}

func primaryColor(img image.Image, smallBucket float64) (uint32, error) {
	colors := color_extractor.ExtractColorsWithConfig(img, color_extractor.Config{
		SmallBucket: smallBucket,
		DownSizeTo:  224,
	})
	if len(colors) == 0 {
		return 0, ErrNoPrimaryColor
	}
	return colorRGB(colors[0]), nil
}

func colorRGB(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	return (r << 16) | (g << 8) | b
}
