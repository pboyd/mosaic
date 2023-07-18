package mosaic

import (
	"math"
)

// hueResolution is the number of hue buckets to use.
const hueResolution = 100

// ImageList indexes images by color.
type ImageList struct {
	items map[uint32][]string
}

func newImageList() *ImageList {
	return &ImageList{
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
}

// FindNearest returns the nearest color and the images associated with it.
func (il *ImageList) FindNearest(color uint32) (uint32, []string) {
	var nearestColor uint32
	var nearestDistance uint32 = math.MaxUint32
	for c := range il.items {
		distance := colorDistance(color, c)
		if distance < nearestDistance {
			nearestColor = c
			nearestDistance = distance
		}
	}

	return nearestColor, il.items[nearestColor]
}
