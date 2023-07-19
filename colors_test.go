package mosaic

import (
	"image"
	"testing"

	_ "image/png"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestAverageColor(t *testing.T) {
	assert := assert.New(t)
	img := loadTestImage(t, "testfiles/two-tone.png")
	c := averageColor(img)
	assert.Equal(uint32(0x7f7f7f), c)
}

func TestPrimaryColor(t *testing.T) {
	assert := assert.New(t)
	img := loadTestImage(t, "testfiles/cat.jpg")
	crop := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(10, 10, 20, 20))
	spew.Dump(crop.Bounds())
	c := primaryColor(crop)
	assert.Equal(uint32(0x000000), c)
}
