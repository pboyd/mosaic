package mosaic

import (
	"image"
	"os"
	"testing"
)

func loadTestImage(t *testing.T, path string) image.Image {
	fh, err := os.Open(path)
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
