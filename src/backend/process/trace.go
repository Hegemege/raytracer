package process

import (
	"image/color"
	"math"
	"math/rand"
	"raytracer/models"

	"github.com/go-gl/mathgl/mgl32"
)

func Trace(context *models.RenderContext, ray *models.Ray) mgl32.Vec3 {
	// Distance to hit, can be used to create a depth map too
	var t float32 = math.MaxFloat32

	// If no hits, return background color
	//r := 256 * float32(ray.X) / float32(context.Width)
	//g := 256 * float32(ray.Y) / float32(context.Height)
	//b := 128
	r := float32(255.0)
	g := float32(255.0)
	b := float32(255.0)
	c := mgl32.Vec3{
		r / 255.0,
		g / 255.0,
		b / 255.0,
	}

	hit := false
	bounce := false
	var hitNormal mgl32.Vec3

	for _, sphere := range context.Scene.Spheres {
		distance := sphere.RayIntersect(ray)
		if distance > 0 && distance < t {
			hit = true
			bounce = true
			t = distance
			normal := sphere.NormalAt(ray.Origin.Add(ray.Direction.Mul(distance)))
			hitNormal = normal
			c = mgl32.Vec3{
				(normal.X() + 1) * 0.5,
				(normal.Y() + 1) * 0.5,
				(normal.Z() + 1) * 0.5,
			}
		}
	}

	for _, triangle := range context.Triangles {
		distance := triangle.RayIntersect(ray)
		if distance > 0 && distance < t {
			hit = true
			bounce = true
			// Don't bounce from lights
			if triangle.Material.Name == "Light" {
				bounce = false
			}

			hitNormal = triangle.Normal // TODO: Smoothing? Average vertex normals by u,v

			t = distance
			r := triangle.Material.Kd[0]
			g := triangle.Material.Kd[1]
			b := triangle.Material.Kd[2]
			c = mgl32.Vec3{r, g, b}
		}
	}

	// Bounces
	if hit && bounce && ray.Bounce < context.BounceLimit {
		hitPoint := ray.Origin.Add(ray.Direction.Mul(t))
		gatheredColor := mgl32.Vec3{0, 0, 0}
		for i := 0; i < context.BounceRays; i++ {
			// Sample random in unit sphere and normalize
			sample := randomInHemisphere(hitNormal).Normalize()

			// Spawn new ray
			bounceRay := models.Ray{
				Origin:    hitPoint, //.Add(sample.Mul(0.01)),
				Bounce:    ray.Bounce + 1,
				Direction: sample,
				X:         ray.X,
				Y:         ray.Y,
			}

			bounceColor := Trace(context, &bounceRay)
			gatheredColor = gatheredColor.Add(bounceColor)
		}

		totalColor := gatheredColor.Mul(1.0 / float32(context.BounceRays))

		c = multiplyColor(c, totalColor)
	}

	return c

}

func randomInHemisphere(normal mgl32.Vec3) mgl32.Vec3 {
	inUnitSphere := randomInUnitSphere()
	if inUnitSphere.Dot(normal) > 0.0 {
		return inUnitSphere
	}

	return inUnitSphere.Mul(-1)
}

func randomInUnitSphere() mgl32.Vec3 {
	for {
		p := mgl32.Vec3{rand.Float32()*2 - 1, rand.Float32()*2 - 1, rand.Float32()*2 - 1}
		if p.LenSqr() < 1 {
			return p
		}
	}
}

func multiplyColor(c1 mgl32.Vec3, c2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		c1[0] * c2[0],
		c1[1] * c2[1],
		c1[2] * c2[2],
	}
}

func addColor(c1 color.RGBA, c2 color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(math.Max((float64(c1.R)/255.0)+(float64(c2.R)/255.0), 0.0), 1.0) * 255),
		G: uint8(math.Min(math.Max((float64(c1.G)/255.0)+(float64(c2.G)/255.0), 0.0), 1.0) * 255),
		B: uint8(math.Min(math.Max((float64(c1.B)/255.0)+(float64(c2.B)/255.0), 0.0), 1.0) * 255),
		A: 255,
	}
}
