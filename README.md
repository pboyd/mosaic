# Mosaic

Generates images from many small tiles.

This is still in an early stage of development, but it is currently functional. Here's an example:

[![Go Gopher](https://raw.githubusercontent.com/pboyd/mosaic/master/examples/gopher.small.png)](https://raw.githubusercontent.com/pboyd/mosaic/master/examples/gopher.png)


## Usage

- Ensure a recent version of Go is installed.
- Clone this repo
- Inside this project's directory run: `go build ./cmd/mosaic`


```
./mosaic -image source.jpg -out output.jpg -tiles path/to/tile/images
```

## Credit

For the example image:
- The Go Gopher was designed by [Renee French](http://reneefrench.blogspot.com/)
- Base image by [Takuya Ueda](https://github.com/golang-samples/gopher-vector)
- Tile images were gathered from around the internet, but mainly the [Go blog](https://go.dev/blog/gopher)
