package models

import (
	"bytes"
	"image"

	"image/color"
	"image/draw"
	_ "image/png"

	"github.com/go-gl/mathgl/mgl32"
)

type Texture struct {
	Name string

	// Parsed
	Buffer  []byte
	Texture *image.RGBA
	Width   int
	Height  int
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
	switch img.ColorModel() {
	case color.RGBAModel:
		t.Texture = img.(*image.RGBA)
	case color.NRGBAModel:
		rawImage := img.(*image.NRGBA)
		bounds := rawImage.Bounds()
		t.Texture = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(t.Texture, t.Texture.Bounds(), rawImage, bounds.Min, draw.Src)
	}

	return &t
}

func (t *Texture) SampleUV(uv mgl32.Vec2) mgl32.Vec3 {
	coords := t.GetTexelCoordinates(uv)
	r, g, b, _ := t.Texture.At(int(coords.X()), int(coords.Y())).RGBA() // RGBA returns up to 0xFFFF = 65535
	return mgl32.Vec3{float32(r), float32(g), float32(b)}.Mul(1.0 / 65535.0)
}

func (t *Texture) GetTexelCoordinates(uv mgl32.Vec2) mgl32.Vec2 {
	return mgl32.Vec2{uv.X() * float32(t.Width), uv.Y() * float32(t.Height)}
}
