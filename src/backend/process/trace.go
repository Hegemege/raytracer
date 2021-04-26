package process

import (
	"image/color"
	"raytracer/models"
)

func Trace(context *models.RenderContext, ray *models.Ray) color.Color {

	for _, sphere := range context.Scene.Spheres {
		distance := sphere.RayIntersect(ray)
		if distance > 0 {
			normal := sphere.NormalAt(ray.Origin.Add(ray.Direction.Mul(distance)))
			return color.RGBA{
				R: uint8((normal.X() + 1) * 128),
				G: uint8((normal.Y() + 1) * 128),
				B: uint8((normal.Z() + 1) * 128),
				A: 255,
			}
			//return sphere.Material.Albedo
		}
	}

	// If no hits, return background color
	r := 256 * float32(ray.X) / float32(context.Width)
	g := 256 * float32(ray.Y) / float32(context.Height)
	b := 128
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
