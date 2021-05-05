package models

import (
	"bytes"
	"image"

	_ "image/png"
)

type Texture struct {
	Name string

	// Parsed
	Buffer []byte
	Image  *image.Image
	Width  int
	Height int
}

func NewTexture(name string, buffer *[]byte) *Texture {
	t := Texture{
		Name: name,
	}

	reader := bytes.NewReader(*buffer)

	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	t.Width = img.Bounds().Dx()
	t.Height = img.Bounds().Dy()
	t.Image = &img

	return &t
}
