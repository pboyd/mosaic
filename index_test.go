package mosaic

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type indexTestCase struct {
	color    uint32
	exact    bool
	expected []string
}

var testfilesCases = []indexTestCase{
	{0x0000FF, true, []string{"testfiles/blue.png"}},
	{0x00FF00, true, []string{"testfiles/green.png"}},
	{0xFF0000, true, []string{"testfiles/red.png"}},
	{0xFFFFFF, true, []string{"testfiles/white.png"}},
	{0x0011DD, false, []string{"testfiles/blue.png"}},
	{0xFF33FF, false, []string{"testfiles/pink.png"}},
	{0xAAAA12, false, []string{"testfiles/yellow.png"}},
	{0x010101, false, []string{"testfiles/black.png", "testfiles/two-tone.png"}},
	{0x777777, false, []string{"testfiles/gray.png"}},
}

func TestIndex(t *testing.T) {
	assert := assert.New(t)
	index := NewIndex(Config{})
	err := index.AddPath(context.Background(), "testfiles")
	if !assert.NoError(err) || !assert.Greater(index.Len(), 0) {
		return
	}

	checkIndex(assert, index, testfilesCases)
}

func TestIndexSerialize(t *testing.T) {
	assert := assert.New(t)
	index := NewIndex(Config{})
	err := index.AddPath(context.Background(), "testfiles")
	if !assert.NoError(err) || !assert.Greater(index.Len(), 0) {
		return
	}

	buf := new(bytes.Buffer)
	err = index.Write(buf)
	if !assert.NoError(err) {
		return
	}

	newIndex := NewIndex(Config{})
	err = newIndex.Load(buf)
	if !assert.NoError(err) {
		return
	}

	checkIndex(assert, newIndex, testfilesCases)

	assert.Equal(index.Len(), newIndex.Len())
	for color, paths := range index.paths {
		assert.ElementsMatch(paths, newIndex.paths[color])
	}
}

func checkIndex(assert *assert.Assertions, index *Index, cases []indexTestCase) {
	for i, c := range cases {
		color, paths := index.FindNearest(c.color)
		if c.exact {
			assert.Equal(c.color, color, "case %d", i)
		}
		assert.Equal(c.expected, paths, "case %d", i)
	}
}
