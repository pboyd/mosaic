package mosaic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageList(t *testing.T) {
	cases := []struct {
		color    uint32
		exact    bool
		expected []string
	}{
		{0x0000FF, true, []string{"testfiles/blue.png"}},
		{0x00FF00, true, []string{"testfiles/green.png"}},
		{0xFF0000, true, []string{"testfiles/red.png"}},
		{0xFFFFFF, true, []string{"testfiles/white.png"}},
		{0x0011DD, false, []string{"testfiles/blue.png"}},
		{0xFFAAFF, false, []string{"testfiles/pink.png"}},
		{0xC19D25, false, []string{"testfiles/yellow.png"}},
	}

	assert := assert.New(t)
	list := BuildImageList("testfiles")
	if !assert.Greater(list.Len(), 0) {
		return
	}

	for i, c := range cases {
		color, paths := list.FindNearest(c.color)
		if c.exact {
			assert.Equal(c.color, color, "case %d", i)
		}
		assert.Equal(c.expected, paths, "case %d", i)
	}
}
