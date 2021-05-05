package process

import (
	"math"
	"raytracer/models"
	"raytracer/utility"

	"github.com/go-gl/mathgl/mgl32"
)

type RaycastResult struct {
	Triangle *models.Triangle
	T        float32
	U        float32
	V        float32
	Point    mgl32.Vec3
}

// Path traces a given pixel ray
func Trace(context *models.RenderContext, ray *models.Ray) mgl32.Vec3 {
	// If out of scene, return background color
	// "Ambient color"
	r := float32(0.0)
	g := float32(0.0)
	b := float32(0.0)
	c := mgl32.Vec3{
		r / 255.0,
		g / 255.0,
		b / 255.0,
	}

	result := rayCast(context, ray)

	shadingTerms := make([]mgl32.Vec3, 0)
	brdfTerms := make([]mgl32.Vec3, 0)

	indirectCounter := 0

	currentDir := ray.Direction

	for {
		if result == nil {
			break
		}

		shading := mgl32.Vec3{0, 0, 0}
		diffuse, normal, _ := getMaterialParameters(context, result)

		lightSampleN := 25
		for i := 0; i < lightSampleN; i++ {
			lightSample, pdf := context.Light.Sample()

			shadowRay := lightSample.Sub(result.Point)
			shadowRayN := shadowRay.Normalize()
			lightIncident := shadowRayN.Dot(context.Light.Normal)
			if lightIncident < 0 {
				sRay := &models.Ray{
					Origin:    result.Point,
					Direction: shadowRayN,
					X:         ray.X,
					Y:         ray.Y,
					Bounce:    ray.Bounce,
				}
				shadowResult := rayCast(context, sRay)

				if shadowResult != nil && shadowResult.Triangle.Material.Name == "Light" {
					theta_l := float32(math.Max(float64(-lightIncident), 0.0))
					theta := float32(math.Max(float64(shadowRayN.Dot(result.Triangle.Normal)), 0.0))
					radius2 := shadowRay.LenSqr()

					color := utility.MultiplyColor(diffuse, context.Light.Emission).Mul(theta_l * theta / (radius2 * pdf * math.Pi))
					color = utility.ClampColor(color)

					shading = shading.Add(color)
				}
			}
		}

		shading = shading.Mul(1 / float32(lightSampleN))

		shadingTerms = append(shadingTerms, shading)

		if indirectCounter >= int(context.BounceLimit) {
			brdfTerms = append(brdfTerms, mgl32.Vec3{0, 0, 0})
			break
		}

		// Sample from hemisphere
		sample := utility.RandomInHemisphere(result.Triangle.Normal).Normalize()

		bounceRay := models.Ray{
			Origin:    result.Point, //.Add(sample.Mul(0.01)),
			Bounce:    ray.Bounce + 1,
			Direction: sample,
			X:         ray.X,
			Y:         ray.Y,
		}

		// New bounce
		result = rayCast(context, &bounceRay)
		indirectCounter++

		brdfTheta := currentDir.Dot(sample) * -1
		theta := sample.Dot(normal)

		pdf := math.Cos(float64(brdfTheta)) / math.Pi
		brdfTerm := diffuse.Mul(float32(math.Cos(float64(theta)) / (math.Pi * pdf)))

		brdfTerms = append(brdfTerms, brdfTerm)

		currentDir = sample
	}

	// calculate E_i from back of the values to front
	// e.g. shading_1 + BRDF_1*(shading_2 + BRDF_2*(shading_3 + BRDF_3)))

	for i := len(shadingTerms) - 2; i >= 0; i-- {
		brdfTerms[i] = utility.MultiplyColor(brdfTerms[i], shadingTerms[i+1].Add(brdfTerms[i+1]))
	}

	if len(shadingTerms) > 0 {
		return shadingTerms[0].Add(brdfTerms[0])
	}

	// Did not hit anything
	return c
}

func rayCast(context *models.RenderContext, ray *models.Ray) *RaycastResult {
	context.Rays += 1

	// Distance to hit, can be used to create a depth map too
	var tmin float32 = math.MaxFloat32
	var umin float32 = 0.0
	var vmin float32 = 0.0
	var imin int // Triangle index

	/*
		for _, sphere := range context.Scene.Spheres {
			t := sphere.RayIntersect(ray)
			if t > 0 && t < tmin {
				tmin = t
			}
		}
	*/

	for i, triangle := range context.Triangles {
		if triangle.Normal.Dot(ray.Direction) > 0 {
			continue
		}
		t, u, v := triangle.RayIntersect(ray)
		if t > 0 && t < tmin {
			tmin = t
			umin = u
			vmin = v
			imin = i
		}
	}

	if tmin < math.MaxFloat32 {
		return &RaycastResult{
			Triangle: &context.Triangles[imin],
			T:        tmin,
			U:        umin,
			V:        vmin,
			Point:    ray.Origin.Add(ray.Direction.Mul(tmin)),
		}
	}

	return nil
}

func getMaterialParameters(context *models.RenderContext, result *RaycastResult) (diffuse mgl32.Vec3, normal mgl32.Vec3, specular mgl32.Vec3) {
	r := result.Triangle.Material.Kd[0]
	g := result.Triangle.Material.Kd[1]
	b := result.Triangle.Material.Kd[2]
	diffuse = mgl32.Vec3{r, g, b}
	normal = result.Triangle.Normal

	// TODO: Normal map, specular map sampling
	// TODO: Texture sampling

	return diffuse, normal, specular
}
