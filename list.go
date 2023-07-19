package mosaic

import (
	"log"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/cluster"
)

// hueResolution is the number of hue buckets to use.
const hueResolution = 100

// ImageList indexes images by color.
type ImageList struct {
	model *cluster.KNN

	items        map[uint32][]string
	trainingData [][]float64
	expected     []float64
}

func newImageList() *ImageList {
	return &ImageList{
		model: cluster.NewKNN(1, nil, nil, base.EuclideanDistance),
		items: map[uint32][]string{},
	}
}

// Len returns the number of items in the list.
func (il *ImageList) Len() int {
	return len(il.items)
}

func (il *ImageList) insert(color uint32, path string) {
	if _, exists := il.items[color]; exists {
		il.items[color] = append(il.items[color], path)
		return
	}
	il.items[color] = []string{path}

	il.trainingData = append(il.trainingData, colorVector(color))
	il.expected = append(il.expected, float64(color))
	il.model.UpdateTrainingSet(il.trainingData, il.expected)
}

// FindNearest returns the nearest color and the images associated with it.
func (il *ImageList) FindNearest(color uint32) (uint32, []string) {
	guess, err := il.model.Predict(colorVector(color))
	if err != nil {
		log.Printf("error finding nearest color: %v", err)
		return 0, nil
	}

	nearestColor := uint32(guess[0])
	paths, ok := il.items[nearestColor]
	if !ok {
		log.Printf("no images found for color %v", nearestColor)
		return 0, nil
	}
	return nearestColor, paths
}
