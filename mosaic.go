package mosaic

import (
	"image"
	"log"
	"math/rand"
	"os"

	"github.com/disintegration/imaging"
	"golang.org/x/image/draw"
)

// Config contains options for mosaic generation.
type Config struct {
	TileSize int
}

// Generate generates a mosaic image from the source image and the tile images.
func Generate(src image.Image, tileImages *ImageList, config Config) image.Image {
	output := image.NewRGBA(src.Bounds())
	for tile := range tileize(src, config.TileSize) {
		c := primaryColor(tile)

		_, paths := tileImages.FindNearest(c)
		log.Printf("(%d, %d) - %.6x -> %v", tile.Bounds().Min.X, tile.Bounds().Min.Y, c, paths)
		replacement, err := loadImage(pickOne(paths))
		if err != nil {
			// This shouldn't happen, because we already loaded all the images.
			log.Printf("Failed to load image: %v", err)
			continue
		}

		replacement = imaging.Fill(replacement, config.TileSize, config.TileSize, imaging.Center, imaging.Lanczos)
		draw.Copy(output, tile.Bounds().Min, replacement, replacement.Bounds(), draw.Src, nil)
	}
	return output
}

// tileize divides the source image into tiles.
func tileize(src image.Image, tileSize int) <-chan image.Image {
	ch := make(chan image.Image)
	go func() {
		defer close(ch)

		bounds := src.Bounds()
		for x := bounds.Min.X; x < bounds.Max.X; x += tileSize {
			for y := bounds.Min.Y; y < bounds.Max.Y; y += tileSize {
				r := image.Rect(x, y, x+tileSize, y+tileSize)
				if r.Max.X > bounds.Max.X {
					r.Max.X = bounds.Max.X
				}
				if r.Max.Y > bounds.Max.Y {
					r.Max.Y = bounds.Max.Y
				}
				si := subImage(src, r)
				ch <- si
			}
		}
	}()

	return ch
}

func subImage(src image.Image, rect image.Rectangle) image.Image {
	si, ok := src.(subImager)
	if !ok {
		panic("unsupported image type")
	}
	return si.SubImage(rect)
}

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

func pickOne(s []string) string {
	return s[rand.Intn(len(s))]
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return imaging.Decode(f)
}
