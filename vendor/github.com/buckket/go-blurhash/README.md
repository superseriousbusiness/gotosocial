# go-blurhash [![Build Status](https://travis-ci.org/buckket/go-blurhash.svg)](https://travis-ci.org/buckket/go-blurhash) [![Go Report Card](https://goreportcard.com/badge/github.com/buckket/go-blurhash)](https://goreportcard.com/report/github.com/buckket/go-blurhash) [![codecov](https://codecov.io/gh/buckket/go-blurhash/branch/master/graph/badge.svg)](https://codecov.io/gh/buckket/go-blurhash) [![GoDoc](https://godoc.org/github.com/buckket/go-blurhash?status.svg)](https://pkg.go.dev/github.com/buckket/go-blurhash)

**go-blurhash** is a pure Go implementation of the [BlurHash](https://github.com/woltapp/blurhash) algorithm, which is used by
[Mastodon](https://github.com/tootsuite/mastodon) an other Fediverse software to implement a swift way of preloading placeholder images as well
as hiding sensitive media. Read more about it [here](https://blog.joinmastodon.org/2019/05/improving-support-for-adult-content-on-mastodon/).

**tl;dr:** BlurHash is a compact representation of a placeholder for an image.

This library allows generating the BlurHash of a given image, as well as
reconstructing a blurred version with specified dimensions from a given BlurHash.

This library is based on the following reference implementations:
- Encoder: [https://github.com/woltapp/blurhash/blob/master/C](https://github.com/woltapp/blurhash/blob/master/C) (C)
- Deocder: [https://github.com/woltapp/blurhash/blob/master/TypeScript](https://github.com/woltapp/blurhash/blob/master/TypeScript) (TypeScript)

BlurHash is written by [Dag Ågren](https://github.com/DagAgren) / [Wolt](https://github.com/woltapp).

|            | Before                         | After                          |
| ---------- |:------------------------------:| :-----------------------------:|
| **Image**  | ![alt text][test]              | "LFE.@D9F01_2%L%MIVD*9Goe-;WB" |
| **Hash**   | "LFE.@D9F01_2%L%MIVD*9Goe-;WB" | ![alt text][test_blur]

[test]: test.png "Blurhash example input."
[test_blur]: test_blur.png "Blurhash example output"

## Installation

### From source

    go get -u github.com/buckket/go-blurhash

## Usage

go-blurhash exports three functions:
```go
func blurhash.Encode(xComponents, yComponents int, rgba image.Image) (string, error)
func blurhash.Decode(hash string, width, height, punch int) (image.Image, error)
func blurhash.Components(hash string) (xComponents, yComponents int, err error)
```

Here’s a simple demonstration. Check [pkg.go.dev](https://pkg.go.dev/github.com/buckket/go-blurhash) for the full documentation.

```go
package main

import (
	"fmt"
	"github.com/buckket/go-blurhash"
	"image/png"
	"os"
)

func main() {
	// Generate the BlurHash for a given image
	imageFile, _ := os.Open("test.png")
	loadedImage, err := png.Decode(imageFile)
	str, _ := blurhash.Encode(4, 3, loadedImage)
	if err != nil {
		// Handle errors
	}
	fmt.Printf("Hash: %s\n", str)

	// Generate an image for a given BlurHash
	// Width will be 300px and Height will be 500px
	// Punch specifies the contrasts and defaults to 1
	img, err := blurhash.Decode(str, 300, 500, 1)
	if err != nil {
		// Handle errors
	}
	f, _ := os.Create("test_blur.png")
	_ = png.Encode(f, img)
	
	// Get the x and y components used for encoding a given BlurHash
	x, y, err := blurhash.Components("LFE.@D9F01_2%L%MIVD*9Goe-;WB")
	if err != nil {
		// Handle errors
	}
	fmt.Printf("xComponents: %d, yComponents: %d", x, y)
}
```

## Limitations

- Presumably a bit slower than the C implementation (TODO: Benchmarks)

## Notes

- As mentioned [here](https://github.com/woltapp/blurhash#how-fast-is-encoding-decoding), it’s best to
generate very small images (~32x32px) via BlurHash and scale them up to the desired dimensions afterwards for optimal performance.
- Since [#2](https://github.com/buckket/go-blurhash/pull/2) we diverted from the reference implementation by memorizing sRGBtoLinear values, thus increasing encoding speed at the cost of higher memory usage.
- Starting with v1.1.0 the signature of blurhash.Encode() has changed slightly (see [#3](https://github.com/buckket/go-blurhash/issues/3)).

## License

 GNU GPLv3+
 