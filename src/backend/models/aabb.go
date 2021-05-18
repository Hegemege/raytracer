package models

import "github.com/go-gl/mathgl/mgl32"

// Axis-aligned bounding box
type AABB struct {
	Position mgl32.Vec3
	Width    float32
	Height   float32
	Length   float32

	Min mgl32.Vec3
	Max mgl32.Vec3

	// Contains Min and Max for indexing
	Bounds [2]mgl32.Vec3
}

func NewAABBParametric(position mgl32.Vec3, width float32, height float32, length float32) *AABB {
	min := mgl32.Vec3{
		position.X() - width/2.0,
		position.Y() - height/2.0,
		position.Z() - length/2.0,
	}
	max := mgl32.Vec3{
		position.X() + width/2.0,
		position.Y() + height/2.0,
		position.Z() + length/2.0,
	}
	return &AABB{
		Position: position,
		Width:    width,
		Height:   height,
		Length:   length,
		Min:      min,
		Max:      max,
		Bounds:   [2]mgl32.Vec3{min, max},
	}
}

func NewAABBMinMax(min mgl32.Vec3, max mgl32.Vec3) *AABB {
	return &AABB{
		Position: min.Add(max).Mul(0.5),
		Width:    max.X() - min.X(),
		Height:   max.Y() - min.Y(),
		Length:   max.Z() - min.Z(),
		Min:      min,
		Max:      max,
		Bounds:   [2]mgl32.Vec3{min, max},
	}
}

func (aabb *AABB) RayIntersect(ray *Ray) (bool, float32, float32) {
	// From https://www.scratchapixel.com/lessons/3d-basic-rendering/minimal-ray-tracer-rendering-simple-shapes/ray-box-intersection
	var tmin, tmax, tymin, tymax, tzmin, tzmax float32

	tmin = (aabb.Bounds[ray.Sign[0]].X() - ray.Origin.X()) * ray.InvDirection.X()
	tmax = (aabb.Bounds[1-ray.Sign[0]].X() - ray.Origin.X()) * ray.InvDirection.X()
	tymin = (aabb.Bounds[ray.Sign[1]].Y() - ray.Origin.Y()) * ray.InvDirection.Y()
	tymax = (aabb.Bounds[1-ray.Sign[1]].Y() - ray.Origin.Y()) * ray.InvDirection.Y()

	if tmin > tymax || tymin > tmax {
		return false, 0, 0
	}

	if tymin > tmin {
		tmin = tymin
	}

	if tymax < tmax {
		tmax = tymax
	}

	tzmin = (aabb.Bounds[ray.Sign[2]].Z() - ray.Origin.Z()) * ray.InvDirection.Z()
	tzmax = (aabb.Bounds[1-ray.Sign[2]].Z() - ray.Origin.Z()) * ray.InvDirection.Z()

	if tmin > tzmax || tzmin > tmax {
		return false, 0, 0
	}

	if tzmin > tmin {
		tmin = tzmin
	}

	if tzmax < tmax {
		tmax = tzmax
	}

	return true, tmin, tmax
}
