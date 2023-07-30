package mosaic

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/disintegration/imaging"
)

func BenchmarkIndexLarge(b *testing.B) {
	tmp := makeTestImages(b, 0, 0)
	b.ResetTimer()

	cfg := Config{
		Workers: runtime.GOMAXPROCS(0),
	}

	err := NewIndex(cfg).AddPath(context.Background(), tmp)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkIndexTiny(b *testing.B) {
	tmp := makeTestImages(b, 100, 100)
	b.ResetTimer()

	cfg := Config{
		Workers: runtime.GOMAXPROCS(0),
	}
	err := NewIndex(cfg).AddPath(context.Background(), tmp)
	if err != nil {
		b.Fatal(err)
	}
}

// makeTestImages creates b.N test images in a temporary directory.
// If width and height are non-zero, then the images will be resized.
// A cleanup handler is registered to remove the temporary directory.
func makeTestImages(b *testing.B, width, height int) string {
	src := loadTestImage(b, "testfiles/cat.jpg")

	if width != 0 || height != 0 {
		src = imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)
	}

	tmp, err := os.MkdirTemp("", "mosaic-benchmark")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	for i := 0; i < b.N; i++ {
		func() {
			out, err := os.Create(filepath.Join(tmp, fmt.Sprintf("%d.jpg", i)))
			if err != nil {
				b.Fatal(err)
			}
			defer out.Close()

			err = imaging.Encode(out, src, imaging.JPEG)
			if err != nil {
				b.Fatal(err)
			}
		}()
	}

	return tmp
}

func BenchmarkSwap(b *testing.B) {
	cfg := Config{
		Workers: runtime.GOMAXPROCS(0),
	}

	idx := benchmarkIndex(b)
	/*
		idx := NewIndex(cfg)
		err := idx.AddPath(context.Background(), "/home/user/lorempicsum")
		if err != nil {
			b.Fatal(err)
		}
	*/

	src := loadTestImage(b, "testfiles/cat.jpg")

	passes := 1

	// Calculate the area required for each tile to get b.N tiles.
	area := float64(src.Bounds().Dx() * src.Bounds().Dy())
	tileArea := float64(src.Bounds().Dx()*src.Bounds().Dy()) / float64(b.N)

	// If the area is less than 100 pixels, then we need to do multiple passes.
	if tileArea < 50 {
		tileArea = 50
		tileCount := area / tileArea
		passes = b.N / int(math.Ceil(tileCount))
	}

	tileWidth := math.Sqrt(tileArea)
	cfg.TileWidth = int(tileWidth)
	cfg.TileHeight = int(math.Ceil(tileArea / tileWidth))

	//fmt.Printf("passes=%d tileArea=%.02f tileWidth=%d tileHeight=%d\n", passes, tileArea, cfg.TileWidth, cfg.TileHeight)
	//fmt.Printf("n=%d a=%d\n", b.N, passes*int(area/tileArea))

	generator := NewGenerator(cfg, idx)

	b.ResetTimer()

	for i := 0; i < passes; i++ {
		generator.Generate(context.Background(), src)
	}
}

func benchmarkIndex(b *testing.B) *Index {
	// colors is a file containing a list of colors encoded as 32-bit integers.
	buf, err := os.ReadFile("testfiles/colors")
	if err != nil {
		b.Fatal(err)
	}

	index := NewIndex(Config{})
	for i := 0; i < len(buf); i += 4 {
		err = index.insert(binary.LittleEndian.Uint32(buf[i:]), "testfiles/cat.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}

	return index
}
