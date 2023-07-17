package mosaic

import (
	"image"
	"log"
	"os"

	"io/fs"
	"path/filepath"

	color_extractor "github.com/marekm4/color-extractor"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// BuildImageList finds and indexes all images in the given path.
func BuildImageList(path string) *ImageList {
	imageCh := findImages(path)
	imageList := &ImageList{}
	for img := range imageCh {
		colors := color_extractor.ExtractColors(img.Image)

		/*
			fmt.Printf("found image: %s\n", img.Path)
			for _, color := range colors {
				fmt.Printf("  #%.6x\n", colorRGB(color))
			}
		*/

		imageList.insert(colorRGB(colors[0]), img.Path)
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

			fh, err := os.Open(path)
			if err != nil {
				log.Printf("%s: %s", path, err)
				return nil
			}
			defer fh.Close()

			img, _, err := image.Decode(fh)
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
