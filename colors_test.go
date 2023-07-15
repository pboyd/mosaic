package mosaic

import (
	"image"
	"os"
	"testing"

	"image/color"
	_ "image/png"

	"github.com/stretchr/testify/assert"
)

func TestAverageColor(t *testing.T) {
	assert := assert.New(t)
	img := loadImage(t, "testfiles/two-tone.png")
	c := averageColor(img).(color.RGBA)
	assert.Equal(uint8(0x7f), c.R)
	assert.Equal(uint8(0x7f), c.G)
	assert.Equal(uint8(0x7f), c.B)
	assert.Equal(uint8(0xff), c.A)
}

func loadImage(t *testing.T, path string) image.Image {
	fh, err := os.Open("testfiles/two-tone.png")
	if err != nil {
		t.Fatal(err)
	}

	defer fh.Close()
	img, _, err := image.Decode(fh)
	if err != nil {
		t.Fatal(err)
	}

	return img
}
