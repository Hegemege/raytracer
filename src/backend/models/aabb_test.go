package models

import (
	"raytracer/utility"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestRayAABBIntersection(t *testing.T) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)

	ray := NewRay(mgl32.Vec3{
		1, 0, 0,
	}, mgl32.Vec3{
		-1, 0, 0,
	}, 0, 0, 0)

	if hit, _, _ := aabb.RayIntersect(ray); !hit {
		t.Errorf("Ray-AABB intersection not working")
	}
}

func TestRayAABBIntersectionEdge(t *testing.T) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)

	ray := NewRay(mgl32.Vec3{
		1, 0.5, 0,
	}, mgl32.Vec3{
		-1, 0, 0,
	}, 0, 0, 0)

	if hit, _, _ := aabb.RayIntersect(ray); !hit {
		t.Errorf("Ray-AABB intersection not working")
	}
}

func TestRayAABBIntersectionEdgeShouldMiss(t *testing.T) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)

	ray := NewRay(mgl32.Vec3{
		1, 0.501, 0,
	}, mgl32.Vec3{
		-1, 0, 0,
	}, 0, 0, 0)

	if hit, _, _ := aabb.RayIntersect(ray); hit {
		t.Errorf("Ray-AABB intersection not working")
	}
}

func TestRayAABBIntersectionAllSides(t *testing.T) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)

	axes := [6]mgl32.Vec3{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
		{-1, 0, 0},
		{0, -1, 0},
		{0, 0, -1},
	}
	for _, axis := range axes {
		ray := NewRay(axis, axis.Mul(-1), 0, 0, 0)

		if hit, _, _ := aabb.RayIntersect(ray); !hit {
			t.Errorf("Ray-AABB intersection not working ray orig %v dir %v", ray.Origin, ray.Direction)
		}
	}
}

func TestRayAABBDistance(t *testing.T) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)

	ray := NewRay(mgl32.Vec3{
		1, 0, 0,
	}, mgl32.Vec3{
		-1, 0, 0,
	}, 0, 0, 0)

	if hit, tmin, tmax := aabb.RayIntersect(ray); !hit || tmin != 0.5 || tmax != 1.5 {
		t.Errorf("Ray-AABB intersection not working")
	}
}

func BenchmarkRayAABB(b *testing.B) {
	aabb := NewAABBParametric(mgl32.Vec3{0, 0, 0}, 1.0, 1.0, 1.0)
	// Create a random ray
	origin := utility.RandomInUnitSphere().Normalize()
	direction := utility.RandomInUnitSphere().Normalize()
	ray := NewRay(origin, direction, 0, 0, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aabb.RayIntersect(ray)
	}
}
