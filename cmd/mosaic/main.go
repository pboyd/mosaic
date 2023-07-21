package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pboyd/mosaic"

	"image/gif"
	"image/jpeg"
	"image/png"
)

var tileImagesPath string
var sourceImagePath string
var outputImagePath string

var config = mosaic.Config{
	ResizeTiles: true,
}

func init() {
	var tileSize int

	flag.StringVar(&sourceImagePath, "image", "", "Path to source image")
	flag.StringVar(&tileImagesPath, "tiles", "", "Path to directory of tile images")
	flag.StringVar(&outputImagePath, "out", "", "Path to output image")
	flag.IntVar(&tileSize, "size", 10, "Tile size")
	flag.IntVar(&config.TileWidth, "tile-width", 0, "Tile width")
	flag.IntVar(&config.TileHeight, "tile-height", 0, "Tile height")
	flag.IntVar(&config.Workers, "workers", runtime.NumCPU(), "Number of workers")
	flag.BoolVar(&config.Blend, "blend", false, "For transparent images, blend the tile images onto the source image")
	flag.Float64Var(&config.Scale, "scale", 1.0, "Scale the source image by this factor")
	flag.Parse()

	if sourceImagePath == "" {
		fmt.Fprintln(os.Stderr, "Missing source image path")
		os.Exit(1)
	}

	if tileImagesPath == "" {
		fmt.Fprintln(os.Stderr, "Missing tile images path")
		os.Exit(1)
	}

	if outputImagePath == "" {
		outputImagePath = deriveOutputImagePath(sourceImagePath)
	}

	if config.TileWidth == 0 {
		config.TileWidth = tileSize
	}
	if config.TileHeight == 0 {
		config.TileHeight = tileSize
	}
	if config.TileWidth <= 0 || config.TileHeight <= 0 {
		fmt.Fprintln(os.Stderr, "Invalid tile size")
		os.Exit(1)
	}

	ext := filepath.Ext(outputImagePath)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
	default:
		fmt.Fprintln(os.Stderr, "Unknown output image type")
		os.Exit(1)
	}
}

func main() {
	sourceImage, err := loadImage(sourceImagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	outFile, err := os.Create(outputImagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer outFile.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()
	tileImages := mosaic.BuildImageList(ctx, tileImagesPath, config)
	fmt.Printf("Loaded %d tile images in %s\n", tileImages.Len(), time.Since(start))

	outputImage := mosaic.Generate(ctx, sourceImage, tileImages, config)

	err = writeImage(outputImage, outFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func deriveOutputImagePath(sourceImagePath string) string {
	dir, base := filepath.Split(sourceImagePath)
	ext := filepath.Ext(sourceImagePath)
	base = strings.TrimSuffix(base, ext)

	outPath := filepath.Join(dir, base+".mosaic"+ext)
	if _, err := os.Stat(outPath); err != nil {
		return outPath
	}

	for i := 2; ; i++ {
		outPath = filepath.Join(dir, base+".mosaic"+fmt.Sprintf("%d", i)+ext)
		if _, err := os.Stat(outPath); err != nil {
			return outPath
		}
	}
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %v", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding %s: %v", path, err)
	}

	return img, nil
}

func writeImage(img image.Image, f *os.File) error {
	ext := filepath.Ext(f.Name())
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, nil)
	case ".png":
		return png.Encode(f, img)
	case ".gif":
		return gif.Encode(f, img, nil)
	}

	return fmt.Errorf("unknown image type: %s", ext)
}
