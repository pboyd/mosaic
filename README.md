# Mosaic

Generates images from many small tiles.

Here's an example:

[![Go Gopher](https://raw.githubusercontent.com/pboyd/mosaic/master/examples/gopher.small.png)](https://raw.githubusercontent.com/pboyd/mosaic/master/examples/gopher.png)

(click to enlarge)

This was generated with:

```
./mosaic -image gopher.png -tiles gophers/ -tile-width 50 -tile-height 70 -blend
```

## Usage

- Ensure a recent version of Go is installed.
- Clone this repo
- Inside this project's directory run: `go build ./cmd/mosaic`

```
./mosaic -image source.jpg -out output.jpg -tiles path/to/tile/images
```

## Credit

In addition to the Go standard library, this program also uses the following modules. Many thanks to their authors.

- [github.com/marekm4/color-extractor](https://github.com/marekm4/color-extractor) finds the primary color of images and tiles.
- The KNN implementation from [github.com/cdipaolo/goml](https://github.com/cdipaolo/goml) finds the closest match for a tile color.
- [github.com/disintegration/imaging](github.com/disintegration/imaging) handles all the image resizing.

For the example image:

- The Go Gopher was designed by [Renee French](http://reneefrench.blogspot.com/)
- Base image by [Takuya Ueda](https://github.com/golang-samples/gopher-vector)
- Tile images were gathered from around the internet, but mainly the [Go blog](https://go.dev/blog/gopher)
