package mosaic

import (
	"math"
	"sort"

	"github.com/crazy3lf/colorconv"
)

// hueResolution is the number of hue buckets to use.
const hueResolution = 100

// ImageList indexes images by color.
type ImageList struct {
	// hueIndex is the top-level index, which is sorted by hue.
	hueIndex []*hueBucket
}

// Len returns the number of items in the list.
func (il *ImageList) Len() int {
	total := 0
	for _, hli := range il.hueIndex {
		total += len(hli.items)
	}
	return total
}

func (il *ImageList) insert(color uint32, path string) {
	hue, _, _ := colorconv.RGBToHSV(uint8(color>>16), uint8(color>>8), uint8(color))
	hueInt := int(hue * hueResolution)

	index, exists := il.findHueBucket(hueInt)
	if exists {
		il.hueIndex[index].insert(color, path)
		return
	}

	newItem := &hueBucket{
		hue:   hueInt,
		items: map[uint32][]string{color: []string{path}},
	}

	if index == len(il.hueIndex) {
		il.hueIndex = append(il.hueIndex, newItem)
	} else {
		il.hueIndex = append(il.hueIndex[:index+1], il.hueIndex[index:]...)
		il.hueIndex[index] = newItem
	}
}

// findHueBucket returns the index of the item with the given hue, or the index
// where the item should be inserted if it does not exist.
func (il *ImageList) findHueBucket(h int) (index int, exists bool) {
	if len(il.hueIndex) == 0 {
		return 0, false
	}
	index = sort.Search(len(il.hueIndex), func(i int) bool {
		return il.hueIndex[i].hue >= h
	})

	if index < len(il.hueIndex) {
		bucket := il.hueIndex[index]
		if bucket.hue == h {
			exists = true
		}
	}

	return
}

// FindNearest returns the nearest color and the images associated with it.
func (il *ImageList) FindNearest(color uint32) (uint32, []string) {
	hue, _, _ := colorconv.RGBToHSV(uint8(color>>16), uint8(color>>8), uint8(color))
	hueInt := int(hue * hueResolution)

	index, _ := il.findHueBucket(hueInt)

	var bucket *hueBucket
	if index == 0 {
		bucket = il.hueIndex[index]
	} else if index == len(il.hueIndex) {
		bucket = il.hueIndex[index-1]
	} else {
		// Find the closest hue bucket.
		diffA := hueInt - il.hueIndex[index-1].hue
		diffB := il.hueIndex[index].hue - hueInt
		if diffA < diffB {
			bucket = il.hueIndex[index-1]
		} else {
			bucket = il.hueIndex[index]
		}
	}

	var nearestColor uint32
	var nearestDistance uint32 = math.MaxUint32
	for c := range bucket.items {
		distance := colorDistance(color, c)
		if distance < nearestDistance {
			nearestColor = c
			nearestDistance = distance
		}
	}

	return nearestColor, bucket.items[nearestColor]
}

type hueBucket struct {
	hue   int
	items map[uint32][]string
}

func (hb *hueBucket) insert(color uint32, path string) {
	if _, exists := hb.items[color]; exists {
		hb.items[color] = append(hb.items[color], path)
		return
	}
	hb.items[color] = []string{path}
}
