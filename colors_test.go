package mosaic

import (
	"image"
	"testing"

	_ "image/png"

	"github.com/stretchr/testify/assert"
)

func TestPrimaryColor(t *testing.T) {
	assert := assert.New(t)
	img := loadTestImage(t, "testfiles/cat.jpg")
	crop := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(10, 10, 20, 20))
	c := primaryColor(crop)

	assert.Equal(uint32(0xCDAE94), c)
}
