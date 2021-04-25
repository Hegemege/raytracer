package models

import "math"

type Sphere struct {
	Object

	Radius float32
}

func (sphere *Sphere) RayIntersect(ray *Ray) float32 {
	L := sphere.Position.Sub(ray.Origin)
	tca := L.Dot(ray.Direction)
	if tca < 0 {
		return -1
	}
	d2 := L.Dot(L) - tca*tca
	r2 := sphere.Radius * sphere.Radius
	if d2 > r2 {
		return -1
	}
	thc := float32(math.Sqrt(float64(r2 - d2)))
	t0 := tca - thc
	t1 := tca + thc

	if t0 < 0 {
		return t1
	}
	if t1 < 0 {
		return t0
	}

	if t0 > t1 {
		return t1
	}
	return t0
}
