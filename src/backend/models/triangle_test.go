package models

import (
	"raytracer/utility"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

func BenchmarkRayTriangle(b *testing.B) {
	triangle := NewTriangle(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 0, 0}, mgl32.Vec3{0, 1, 0}, &gwob.Material{}, 0)
	// Create a random ray
	origin := utility.RandomInUnitSphere().Normalize()
	direction := utility.RandomInUnitSphere().Normalize()
	ray := NewRay(origin, direction, 0, 0, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		triangle.RayIntersect(ray)
	}
}
