package process

import (
	"image/color"
	"math"
	"math/rand"
	"raytracer/models"

	"github.com/go-gl/mathgl/mgl32"
)

func Trace(context *models.RenderContext, ray *models.Ray) color.RGBA {
	// Distance to hit, can be used to create a depth map too
	var t float32 = math.MaxFloat32

	// If no hits, return background color
	//r := 256 * float32(ray.X) / float32(context.Width)
	//g := 256 * float32(ray.Y) / float32(context.Height)
	//b := 128
	r := 128
	g := 128
	b := 128
	c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

	hit := false
	bounce := false
	var hitNormal mgl32.Vec3
	var hitAmbientColor color.RGBA

	for _, sphere := range context.Scene.Spheres {
		distance := sphere.RayIntersect(ray)
		if distance > 0 && distance < t {
			hit = true
			bounce = true
			t = distance
			normal := sphere.NormalAt(ray.Origin.Add(ray.Direction.Mul(distance)))
			hitNormal = normal
			hitAmbientColor = color.RGBA{R: 128, G: 128, B: 128, A: 255}
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
			hit = true
			bounce = true
			// Don't bounce from lights
			if triangle.Material.Name == "Light" {
				bounce = false
			}

			hitNormal = triangle.Normal // TODO: Smoothing? Average vertex normals by u,v
			hitAmbientColor = color.RGBA{
				R: uint8(triangle.Material.Ka[0] * 255),
				G: uint8(triangle.Material.Ka[1] * 255),
				B: uint8(triangle.Material.Ka[2] * 255),
				A: 255,
			}

			t = distance
			r := uint8(triangle.Material.Kd[0] * 255)
			g := uint8(triangle.Material.Kd[1] * 255)
			b := uint8(triangle.Material.Kd[2] * 255)
			c = color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	// Bounces
	if hit && bounce && ray.Bounce < context.BounceLimit {
		hitPoint := ray.Origin.Add(ray.Direction.Mul(t))
		gatheredColor := [3]float32{0, 0, 0}
		for i := 0; i < context.BounceRays; i++ {
			// Sample random in unit sphere and normalize
			sample := randomInHemisphere(hitNormal).Normalize()

			// Spawn new ray
			bounceRay := models.Ray{
				Origin:    hitPoint.Add(sample.Mul(0.01)),
				Bounce:    ray.Bounce + 1,
				Direction: sample,
				X:         ray.X,
				Y:         ray.Y,
			}

			cos := hitNormal.Dot(sample)
			bounceColor := Trace(context, &bounceRay)
			gatheredColor[0] += float32(bounceColor.R) * cos
			gatheredColor[1] += float32(bounceColor.G) * cos
			gatheredColor[2] += float32(bounceColor.B) * cos
		}

		bounceColor := color.RGBA{
			R: uint8(
				gatheredColor[0] / float32(context.BounceRays),
			),
			G: uint8(
				gatheredColor[1] / float32(context.BounceRays),
			),
			B: uint8(
				gatheredColor[2] / float32(context.BounceRays),
			),
			A: 255,
		}

		c = multiplyColor(c, bounceColor)
		c = multiplyColor(c, hitAmbientColor)
	}

	// Gamma correction
	gamma := 1.0 / 2.2
	gr := math.Pow(float64(c.R)/255.0, gamma)
	gg := math.Pow(float64(c.G)/255.0, gamma)
	gb := math.Pow(float64(c.B)/255.0, gamma)

	return color.RGBA{
		R: uint8(gr * 255),
		G: uint8(gg * 255),
		B: uint8(gb * 255),
		A: 255,
	}
}

func randomInHemisphere(normal mgl32.Vec3) mgl32.Vec3 {
	inUnitSphere := randomInUnitSphere()
	if inUnitSphere.Dot(normal) > 0.0 {
		return inUnitSphere
	}

	return inUnitSphere.Mul(-1)
}

func randomInUnitSphere() mgl32.Vec3 {
	// rand.Seed(x)
	for {
		p := mgl32.Vec3{rand.Float32()*2 - 1, rand.Float32()*2 - 1, rand.Float32()*2 - 1}
		if p.LenSqr() < 1 {
			return p
		}
	}
}

func multiplyColor(c1 color.RGBA, c2 color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(math.Max((float64(c1.R)/255.0)*(float64(c2.R)/255.0), 0.0), 1.0) * 255),
		G: uint8(math.Min(math.Max((float64(c1.G)/255.0)*(float64(c2.G)/255.0), 0.0), 1.0) * 255),
		B: uint8(math.Min(math.Max((float64(c1.B)/255.0)*(float64(c2.B)/255.0), 0.0), 1.0) * 255),
		A: 255,
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
