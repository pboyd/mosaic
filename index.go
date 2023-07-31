package mosaic

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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
	// StatusHandler is called by AddPath to report the progress of the index operation.
	StatusHandler func(<-chan IndexImage)

	config Config

	model *cluster.KNN

	paths        map[uint32][]string
	trainingData [][]float64
	expected     []float64
}

// IndexImage reports the progress of an index operation.
type IndexImage struct {
	Path  string
	Color uint32
	Err   error
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
		return 0, nil
	}

	nearestColor := uint32(guess[0])
	paths, ok := idx.paths[nearestColor]
	if !ok {
		return 0, nil
	}
	return nearestColor, paths
}

// AddPath finds and indexes images from the given path.
func (idx *Index) AddPath(ctx context.Context, path string) error {
	statusCh := make(chan IndexImage, idx.config.Workers*2)
	statusHandler := idx.defaultStatusHandler
	if idx.StatusHandler != nil {
		statusHandler = idx.StatusHandler
	}
	go statusHandler(statusCh)

	pathCh := idx.findImages(ctx, path)

	colorChs := make([]<-chan IndexImage, idx.config.Workers)
	for i := range colorChs {
		colorChs[i] = idx.worker(pathCh)
	}

	for found := range idx.mergeColorChannels(colorChs...) {
		if found.Err == nil {
			found.Err = idx.insert(found.Color, found.Path)
		}
		statusCh <- found
	}

	if idx.Len() == 0 {
		return ErrNoImagesFound
	}

	return nil
}

func (*Index) defaultStatusHandler(ch <-chan IndexImage) {
	for range ch {
		// do nothing
	}
}

func (idx *Index) insert(color uint32, path string) error {
	if _, exists := idx.paths[color]; exists {
		idx.paths[color] = append(idx.paths[color], path)
		return nil
	}
	idx.paths[color] = []string{path}

	idx.trainingData = append(idx.trainingData, colorVector(color))
	idx.expected = append(idx.expected, float64(color))
	return idx.model.UpdateTrainingSet(idx.trainingData, idx.expected)
}

func (idx *Index) findImages(ctx context.Context, path string) <-chan IndexImage {
	ch := make(chan IndexImage)
	go func() {
		defer close(ch)

		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif":
			default:
				return nil
			}

			ii := IndexImage{
				Path: path,
				Err:  err,
			}

			select {
			case ch <- ii:
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

func (idx *Index) worker(ch <-chan IndexImage) <-chan IndexImage {
	out := make(chan IndexImage)
	go func() {
		defer close(out)

		for ii := range ch {
			if ii.Err == nil {
				ii.Color, ii.Err = idx.processOne(ii.Path)
			}

			out <- ii
		}
	}()

	return out
}

func (idx *Index) processOne(path string) (uint32, error) {
	img, err := loadImage(path)
	if err != nil {
		return 0, err
	}

	if idx.config.ResizeTiles {
		img = imaging.Fill(img, idx.config.TileWidth, idx.config.TileHeight, imaging.Center, imaging.Lanczos)
	}
	return primaryColor(img, idx.config.IndexThreshold)
}

func (idx *Index) mergeColorChannels(chs ...<-chan IndexImage) <-chan IndexImage {
	out := make(chan IndexImage)

	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func(ch <-chan IndexImage) {
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

// Write serializes the index to the given writer.
func (idx *Index) Write(w io.Writer) error {
	var err error
	for color, paths := range idx.paths {
		for _, path := range paths {
			err = binary.Write(w, binary.LittleEndian, color)
			if err != nil {
				return err
			}

			pathLen := len(path)
			if pathLen > math.MaxUint16 {
				return fmt.Errorf("path %s is too long", path)
			}

			err = binary.Write(w, binary.LittleEndian, uint16(pathLen))
			if err != nil {
				return err
			}

			_, err = w.Write([]byte(path))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Load deserializes the index from the given reader.
func (idx *Index) Load(r io.Reader) error {
	var err error
	for err == nil {
		var color uint32
		err = binary.Read(r, binary.LittleEndian, &color)
		if err != nil {
			break
		}

		var pathLen uint16
		err = binary.Read(r, binary.LittleEndian, &pathLen)
		if err != nil {
			break
		}

		path := make([]byte, pathLen)
		_, err = io.ReadFull(r, path)
		if err != nil {
			break
		}

		err = idx.insert(color, string(path))
		if err != nil {
			break
		}
	}

	if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
