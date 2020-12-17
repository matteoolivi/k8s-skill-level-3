package main

import "github.com/matteoolivi/img-blurring-exercise/pkg/imageblur"

func main() {
	// TODO: Make second parameter (path to image where to save the processing) configurable.
	imageblur.Run(make(<-chan struct{}), "/tmp/image-to-blur.jpg")
}
