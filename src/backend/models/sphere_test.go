package models

import (
	"math"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestRaySphereIntersection(t *testing.T) {
	sphere := Sphere{
		Object: Object{
			Position: mgl32.Vec3{0, 0, 5},
		},
		Radius: 2,
	}

	ray := Ray{
		Origin:    mgl32.Vec3{0, 0, 0},
		Direction: mgl32.Vec3{0, 0, 1},
	}

	if sphere.RayIntersect(&ray) < 0 {
		t.Errorf("Ray-sphere intersection not working")
	}
}

func TestRaySphereIntersectionDistance(t *testing.T) {
	sphere := Sphere{
		Object: Object{
			Position: mgl32.Vec3{0, 0, 5},
		},
		Radius: 2,
	}

	ray := Ray{
		Origin:    mgl32.Vec3{0, 0, 0},
		Direction: mgl32.Vec3{0, 0, 1},
	}

	distance := sphere.RayIntersect(&ray)
	if math.Abs(float64(distance-3)) > 0.001 {
		t.Errorf("Ray-sphere intersection not working")
	}
}

func TestRaySphereIntersectionInside(t *testing.T) {
	sphere := Sphere{
		Object: Object{
			Position: mgl32.Vec3{0, 0, 0},
		},
		Radius: 2,
	}

	ray := Ray{
		Origin:    mgl32.Vec3{0, 0, -1},
		Direction: mgl32.Vec3{0, 0, 1},
	}

	distance := sphere.RayIntersect(&ray)
	if distance < 0 {
		t.Errorf("Ray-sphere intersection not working")
	}
	if math.Abs(float64(distance-3)) > 0.001 {
		t.Errorf("Ray-sphere intersection not working")
	}
}

func TestRaySphereIntersectionBehind(t *testing.T) {
	sphere := Sphere{
		Object: Object{
			Position: mgl32.Vec3{0, 0, 0},
		},
		Radius: 2,
	}

	ray := Ray{
		Origin:    mgl32.Vec3{0, 0, 5},
		Direction: mgl32.Vec3{0, 0, 1},
	}

	if sphere.RayIntersect(&ray) > 0 {
		t.Errorf("Ray-sphere intersection not working")
	}
}

func TestRaySphereIntersectionShouldMiss(t *testing.T) {
	sphere := Sphere{
		Object: Object{
			Position: mgl32.Vec3{0, 0, 5},
		},
		Radius: 2,
	}

	ray := Ray{
		Origin:    mgl32.Vec3{0, 0, 0},
		Direction: mgl32.Vec3{0, 1, 0},
	}

	if sphere.RayIntersect(&ray) >= 0 {
		t.Errorf("Ray-sphere intersection not working")
	}
}
