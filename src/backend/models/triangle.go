package models

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

type Triangle struct {
	Vertices [3]mgl32.Vec3
	// TODO: Add vertex normals

	Normal   mgl32.Vec3
	Material *gwob.Material

	// Helpers for trace
	Edge0 mgl32.Vec3
	Edge1 mgl32.Vec3
	Edge2 mgl32.Vec3
}

func (triangle *Triangle) RayIntersect(ray *Ray) float32 {
	// From https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection
	v0v2 := triangle.Edge2.Mul(-1)
	pvec := ray.Direction.Cross(v0v2)
	det := triangle.Edge0.Dot(pvec)
	if det < 0.0001 {
		return -1
	}

	invdet := 1.0 / det
	tvec := ray.Origin.Sub(triangle.Vertices[0])
	u := tvec.Dot(pvec) * invdet
	if u < 0 || u > 1 {
		return -1
	}

	qvec := tvec.Cross(triangle.Edge0)
	v := ray.Direction.Dot(qvec) * invdet
	if v < 0 || u+v > 1 {
		return -1
	}

	t := v0v2.Dot(qvec) * invdet

	return t
}

/*
func (triangle *Triangle) RayIntersect(ray *Ray) float32 {
	// From https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/ray-triangle-intersection-geometric-solution
	nDotRay := triangle.Normal.Dot(ray.Direction)
	if math.Abs(float64(nDotRay)) < 0.0001 {
		return -1 // Parallel
	}

	d := triangle.Normal.Dot(triangle.Vertices[0])
	t := (triangle.Normal.Dot(ray.Origin) + d) / nDotRay
	if t < 0 {
		return -1
	}

	P := ray.Origin.Add(ray.Direction.Mul(t))

	if triangle.Normal.Dot(triangle.Edge0.Cross(P.Sub(triangle.Vertices[0]))) < 0 {
		return -1
	}

	if triangle.Normal.Dot(triangle.Edge1.Cross(P.Sub(triangle.Vertices[1]))) < 0 {
		return -1
	}

	if triangle.Normal.Dot(triangle.Edge2.Cross(P.Sub(triangle.Vertices[2]))) < 0 {
		return -1
	}

	return t
}
*/
