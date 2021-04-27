package process

import (
	"image/color"
	"math"
	"raytracer/models"
)

func Trace(context *models.RenderContext, ray *models.Ray) color.Color {
	// Distance to hit, can be used to create a depth map too
	var t float32 = math.MaxFloat32

	// If no hits, return background color
	r := 256 * float32(ray.X) / float32(context.Width)
	g := 256 * float32(ray.Y) / float32(context.Height)
	b := 128
	c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

	for _, sphere := range context.Scene.Spheres {
		distance := sphere.RayIntersect(ray)
		if distance > 0 && distance < t {
			t = distance
			normal := sphere.NormalAt(ray.Origin.Add(ray.Direction.Mul(distance)))
			c = color.RGBA{
				R: uint8((normal.X() + 1) * 128),
				G: uint8((normal.Y() + 1) * 128),
				B: uint8((normal.Z() + 1) * 128),
				A: 255,
			}
		}
	}

	for _, triangle := range context.Triangles {
		distance := triangle.RayIntersect(ray)
		if distance > 0 && distance < t {
			t = distance
			r := uint8(triangle.Material.Kd[0] * 255)
			g := uint8(triangle.Material.Kd[1] * 255)
			b := uint8(triangle.Material.Kd[2] * 255)
			c = color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	return c
}
