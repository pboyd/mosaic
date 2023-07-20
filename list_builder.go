package mosaic

import (
	"context"
	"log"
	"strings"
	"sync"

	"io/fs"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// BuildImageList finds and indexes all images in the given path.
func BuildImageList(ctx context.Context, workers int, path string) *ImageList {
	pathCh := findImages(ctx, path)

	colorChs := make([]<-chan imageColor, workers)
	for i := range colorChs {
		colorChs[i] = findImageColor(pathCh)
	}

	imageList := newImageList()
	for found := range mergeColorChannels(colorChs...) {
		log.Printf("found image %s", found.Path)
		imageList.insert(found.Color, found.Path)
	}

	return imageList
}

type imageColor struct {
	Path  string
	Color uint32
}

func findImages(ctx context.Context, path string) <-chan string {
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

func findImageColor(ch <-chan string) <-chan imageColor {
	out := make(chan imageColor)
	go func() {
		defer close(out)
		for path := range ch {
			img, err := loadImage(path)
			if err != nil {
				log.Printf("%s: %s", path, err)
				continue
			}

			out <- imageColor{
				Path:  path,
				Color: primaryColor(img),
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
