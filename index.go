package mosaic

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	"io/fs"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/cluster"
	"github.com/disintegration/imaging"
)

// ErrNoImagesFound is returned when no images are found in the given path.
var ErrNoImagesFound = errors.New("no images found")

// Index is a list of images that can be accessed by color.
type Index struct {
	config Config

	model *cluster.KNN

	paths        map[uint32][]string
	trainingData [][]float64
	expected     []float64
}

// NewIndex creates a new, empty, image index.
func NewIndex(config Config) *Index {
	if config.Workers == 0 {
		config.Workers = 1
	}

	return &Index{
		config: config,
		model:  cluster.NewKNN(1, nil, nil, base.EuclideanDistance),
		paths:  map[uint32][]string{},
	}
}

// Len returns the number of items in the list.
func (idx *Index) Len() int {
	return len(idx.paths)
}

// FindNearest returns the nearest color and the images associated with it.
func (idx *Index) FindNearest(color uint32) (uint32, []string) {
	guess, err := idx.model.Predict(colorVector(color))
	if err != nil {
		log.Printf("error finding nearest color: %v", err)
		return 0, nil
	}

	nearestColor := uint32(guess[0])
	paths, ok := idx.paths[nearestColor]
	if !ok {
		log.Printf("no images found for color %v", nearestColor)
		return 0, nil
	}
	return nearestColor, paths
}

// AddPath finds and indexes images from the given path.
func (idx *Index) AddPath(ctx context.Context, path string) error {
	pathCh := idx.findImages(ctx, path)

	colorChs := make([]<-chan imageColor, idx.config.Workers)
	for i := range colorChs {
		colorChs[i] = idx.worker(pathCh)
	}

	var count int
	for found := range mergeColorChannels(colorChs...) {
		log.Printf("found image %s", found.Path)
		idx.insert(found.Color, found.Path)
		count++
	}

	if count == 0 {
		return ErrNoImagesFound
	}

	return nil
}

func (idx *Index) insert(color uint32, path string) {
	if _, exists := idx.paths[color]; exists {
		idx.paths[color] = append(idx.paths[color], path)
		return
	}
	idx.paths[color] = []string{path}

	idx.trainingData = append(idx.trainingData, colorVector(color))
	idx.expected = append(idx.expected, float64(color))
	idx.model.UpdateTrainingSet(idx.trainingData, idx.expected)
}

func (idx *Index) findImages(ctx context.Context, path string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)

		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("%s: %s", path, err)
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif":
			default:
				return nil
			}

			select {
			case ch <- path:
			case <-ctx.Done():
				return fs.SkipAll
			}

			return nil
		})
		if err != nil {
			log.Fatalf("error walking path %s: %s", path, err)
		}
	}()
	return ch
}

type imageColor struct {
	Path  string
	Color uint32
}

func (idx *Index) worker(ch <-chan string) <-chan imageColor {
	out := make(chan imageColor)
	go func() {
		defer close(out)
		for path := range ch {
			img, err := loadImage(path)
			if err != nil {
				log.Printf("%s: %s", path, err)
				continue
			}

			if idx.config.ResizeTiles {
				img = imaging.Fill(img, idx.config.TileWidth, idx.config.TileHeight, imaging.Center, imaging.Lanczos)
			}

			out <- imageColor{
				Path:  path,
				Color: primaryColor(img, 0.01),
			}
		}
	}()

	return out
}

func mergeColorChannels(chs ...<-chan imageColor) <-chan imageColor {
	out := make(chan imageColor)

	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func(ch <-chan imageColor) {
			defer wg.Done()
			for img := range ch {
				out <- img
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
