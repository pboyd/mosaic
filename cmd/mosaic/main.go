package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pboyd/mosaic"
)

var tileImagesPath string
var sourceImagePath string
var outputImagePath string
var tileSize int

func init() {
	flag.StringVar(&sourceImagePath, "image", "", "Path to source image")
	flag.StringVar(&tileImagesPath, "tiles", "", "Path to directory of tile images")
	flag.IntVar(&tileSize, "size", 10, "Tile size")
	flag.StringVar(&outputImagePath, "out", "", "Path to output image")
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
}

func main() {
	images := mosaic.BuildImageList(tileImagesPath)
	_ = images
}

func deriveOutputImagePath(sourceImagePath string) string {
	dir, base := filepath.Split(sourceImagePath)
	ext := filepath.Ext(sourceImagePath)
	base = strings.TrimSuffix(base, ext)

	outPath := filepath.Join(dir, base+".mosaic"+ext)
	if _, err := os.Stat(outPath); err == nil {
		return outPath
	}

	for i := 2; ; i++ {
		outPath = filepath.Join(dir, base+".mosaic"+fmt.Sprintf("%d", i)+ext)
		if _, err := os.Stat(outPath); err != nil {
			return outPath
		}
	}
}
