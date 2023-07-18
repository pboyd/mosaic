package mosaic

import (
	"image"
	"log"

	"io/fs"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// BuildImageList finds and indexes all images in the given path.
func BuildImageList(path string) *ImageList {
	imageCh := findImages(path)
	imageList := newImageList()
	for img := range imageCh {
		log.Printf("found image %s", img.Path)
		imageList.insert(primaryColor(img.Image), img.Path)
	}
	return imageList
}

type foundImage struct {
	Path  string
	Image image.Image
}

func findImages(path string) <-chan foundImage {
	ch := make(chan foundImage)
	go func() {
		defer close(ch)

		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("%s: %s", path, err)
				return nil
			}

			ext := filepath.Ext(path)
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif":
			default:
				return nil
			}

			img, err := loadImage(path)
			if err != nil {
				log.Printf("%s: %s", path, err)
				return nil
			}

			ch <- foundImage{
				Path:  path,
				Image: img,
			}
			return nil
		})
		if err != nil {
			log.Fatalf("error walking path %s: %s", path, err)
		}
	}()
	return ch
}
