package process

import (
	"image/color"
	"raytracer/models"
)

func Trace(context *models.RenderContext, ray *models.Ray) color.Color {
	r := 256 * float32(ray.X) / float32(context.Width-1)
	g := 256 * float32(ray.Y) / float32(context.Height-1)
	b := 128
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
