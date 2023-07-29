package mosaic

import (
	"context"
	"image"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"

	"github.com/disintegration/imaging"
	"golang.org/x/image/draw"
)

// Config contains options for mosaic generation.
type Config struct {
	Workers int
	Blend   bool

	Scale float64

	TileWidth  int
	TileHeight int

	// ResizeTiles indicates whether the tile images should be resized before
	// determining their primary color.
	ResizeTiles bool

	IndexThreshold float64
}

// Generate generates a mosaic image from the source image and the tile images.
func Generate(ctx context.Context, src image.Image, tileImages *Index, config Config) image.Image {
	if config.Scale != 0 {
		src = imaging.Resize(src, int(float64(src.Bounds().Dx())*config.Scale), 0, imaging.Lanczos)
	}

	output := image.NewRGBA(src.Bounds())
	if config.Blend {
		draw.Copy(output, src.Bounds().Min, src, src.Bounds(), draw.Src, nil)
	}

	tiles := tileize(ctx, src, config)

	var wg sync.WaitGroup
	wg.Add(config.Workers)
	for i := 0; i < config.Workers; i++ {
		go func() {
			defer wg.Done()
			matchAndSwapTiles(output, tiles, tileImages, config)
		}()
	}

	wg.Wait()

	return output
}

// tileize divides the source image into tiles.
func tileize(ctx context.Context, src image.Image, config Config) <-chan image.Image {
	ch := make(chan image.Image)
	go func() {
		defer close(ch)

		bounds := src.Bounds()
		for x := bounds.Min.X; x < bounds.Max.X; x += config.TileWidth {
			for y := bounds.Min.Y; y < bounds.Max.Y; y += config.TileHeight {
				r := image.Rect(x, y, x+config.TileWidth, y+config.TileHeight)
				if r.Max.X > bounds.Max.X {
					r.Max.X = bounds.Max.X
				}
				if r.Max.Y > bounds.Max.Y {
					r.Max.Y = bounds.Max.Y
				}

				select {
				case ch <- subImage(src, r):
				case <-ctx.Done():
					return
				}
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

func matchAndSwapTiles(output draw.Image, tiles <-chan image.Image, tileImages *Index, config Config) {
	for tile := range tiles {
		c := primaryColor(tile, 0.01)
		if c == math.MaxUint32 {
			continue
		}

		_, paths := tileImages.FindNearest(c)
		log.Printf("(%d, %d) - %.6x -> %v", tile.Bounds().Min.X, tile.Bounds().Min.Y, c, paths)
		replacement, err := loadImage(pickOne(paths))
		if err != nil {
			// This shouldn't happen, because we already loaded all the images.
			log.Printf("Failed to load image: %v", err)
			continue
		}

		replacement = imaging.Fill(replacement, config.TileWidth, config.TileHeight, imaging.Center, imaging.Lanczos)
		if config.Blend {
			draw.Draw(output, tile.Bounds(), replacement, image.ZP, draw.Over)
		} else {
			draw.Copy(output, tile.Bounds().Min, replacement, replacement.Bounds(), draw.Src, nil)
		}
	}
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return imaging.Decode(f)
}
