package mosaic

import (
	"image"
	"os"
	"testing"

	_ "image/png"

	"github.com/stretchr/testify/assert"
)

func TestAverageColor(t *testing.T) {
	assert := assert.New(t)
	img := loadImage(t, "testfiles/two-tone.png")
	c := averageColor(img)
	assert.Equal(uint32(0x7f7f7f), c)
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
