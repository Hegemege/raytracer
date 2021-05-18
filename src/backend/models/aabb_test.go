package models

import (
	"raytracer/utility"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

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
