package mosaic

import (
	"context"
	"image"
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

// Generator generates a mosaic image from the source image and the tile images.
type Generator struct {
	config        Config
	index         *Index
	StatusHandler func(total int64, progress <-chan GeneratorStatus)
}

// NewGenerator creates a new Generator.
func NewGenerator(config Config, index *Index) *Generator {
	return &Generator{
		config: config,
		index:  index,
	}
}

// GeneratorStatus contains the status of the mosaic generation.
type GeneratorStatus struct {
	Bounds     image.Rectangle
	TileNumber int
	Path       string
	Err        error
}

// Generate generates a mosaic image from the source image and the tile images.
func (g *Generator) Generate(ctx context.Context, src image.Image) image.Image {
	if g.config.Scale != 0 {
		src = imaging.Resize(src, int(float64(src.Bounds().Dx())*g.config.Scale), 0, imaging.Lanczos)
	}

	output := image.NewRGBA(src.Bounds())
	if g.config.Blend {
		draw.Copy(output, src.Bounds().Min, src, src.Bounds(), draw.Src, nil)
	}

	tiles := g.tileize(ctx, src)

	statusChans := make([]<-chan GeneratorStatus, g.config.Workers)
	for i := 0; i < g.config.Workers; i++ {
		statusChans[i] = g.matchAndSwapTiles(output, tiles)
	}

	g.wait(src.Bounds(), statusChans)

	return output
}

func (*Generator) defaultStatusHandler(total int64, ch <-chan GeneratorStatus) {
	for range ch {
		// do nothing
	}
}

// tileize divides the source image into tiles.
func (g *Generator) tileize(ctx context.Context, src image.Image) <-chan image.Image {
	ch := make(chan image.Image)
	go func() {
		defer close(ch)

		bounds := src.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y += g.config.TileHeight {
			for x := bounds.Min.X; x < bounds.Max.X; x += g.config.TileWidth {
				r := image.Rect(x, y, x+g.config.TileWidth, y+g.config.TileHeight)
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

func (g *Generator) matchAndSwapTiles(output draw.Image, tiles <-chan image.Image) <-chan GeneratorStatus {
	out := make(chan GeneratorStatus)
	go func() {
		defer close(out)

		for tile := range tiles {
			status := tileStats(output.Bounds(), tile.Bounds())

			c, err := primaryColor(tile, 0.01)
			if err != nil {
				status.Err = err
				out <- status
				continue
			}

			_, paths := g.index.FindNearest(c)
			path := pickOne(paths)
			status.Path = path

			replacement, err := loadImage(path)
			if err != nil {
				// This shouldn't happen, because we already loaded all the images.
				status.Err = err
				out <- status
				continue
			}

			replacement = imaging.Fill(replacement, g.config.TileWidth, g.config.TileHeight, imaging.Center, imaging.Lanczos)
			if g.config.Blend {
				draw.Draw(output, tile.Bounds(), replacement, image.Point{}, draw.Over)
			} else {
				draw.Copy(output, tile.Bounds().Min, replacement, replacement.Bounds(), draw.Src, nil)
			}
			out <- status
		}
	}()
	return out
}

func tileStats(imageBounds, tileBounds image.Rectangle) GeneratorStatus {
	row := tileBounds.Min.X / tileBounds.Dx()
	col := tileBounds.Min.Y / tileBounds.Dy()
	return GeneratorStatus{
		Bounds:     tileBounds,
		TileNumber: 1 + row + col*(imageBounds.Dx()/tileBounds.Dx()),
	}
}

func (g *Generator) wait(bounds image.Rectangle, statusChans []<-chan GeneratorStatus) {
	statusCh := make(chan GeneratorStatus, g.config.Workers*2)
	statusHandler := g.defaultStatusHandler
	if g.StatusHandler != nil {
		statusHandler = g.StatusHandler
	}

	totalTiles := int64((bounds.Dx() / g.config.TileWidth) * (bounds.Dy() / g.config.TileHeight))
	go statusHandler(totalTiles, statusCh)

	var wg sync.WaitGroup
	wg.Add(len(statusChans))
	for _, ch := range statusChans {
		go func(ch <-chan GeneratorStatus) {
			for s := range ch {
				statusCh <- s
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return imaging.Decode(f)
}
