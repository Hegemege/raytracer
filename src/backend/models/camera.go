package models

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type ProjectionType int

const (
	Perspective ProjectionType = iota
	Ortographic
)

type Camera struct {
	Transform               mgl32.Mat4
	ProjectionPlaneDistance float32
	RaysPerPixel            int
	Projection              ProjectionType

	// Perspective projection
	// Vertical fov
	FieldOfView float32

	// Ortographic projection
	// Half-height of the projection plane
	OrtographicSize float32
}

func (camera *Camera) SpawnRays(resolutionWidth int, resolutionHeight int) []Ray {
	rayCount := resolutionHeight * resolutionWidth * camera.RaysPerPixel
	rays := make([]Ray, rayCount)

	var projectionPlaneTopLeft mgl32.Vec3
	var projectionPlaneBottomRight mgl32.Vec3

	if camera.Projection == Perspective {
		verticalHalfAngle := math.Pi * (camera.FieldOfView / 2.0) / 180.0
		horizontalHalfAngle := verticalHalfAngle * (float32(resolutionWidth) / float32(resolutionHeight))

		forward := mgl32.Vec3{0, 0, camera.ProjectionPlaneDistance}

		left := mgl32.QuatRotate(-horizontalHalfAngle, mgl32.Vec3{0, 1, 0}).Rotate(forward)
		right := mgl32.QuatRotate(horizontalHalfAngle, mgl32.Vec3{0, 1, 0}).Rotate(forward)
		top := mgl32.QuatRotate(-verticalHalfAngle, mgl32.Vec3{1, 0, 0}).Rotate(forward)
		bottom := mgl32.QuatRotate(verticalHalfAngle, mgl32.Vec3{1, 0, 0}).Rotate(forward)

		// Project the vectors onto the projection plane
		left = left.Mul(camera.ProjectionPlaneDistance / left.Z())
		right = right.Mul(camera.ProjectionPlaneDistance / right.Z())
		top = top.Mul(camera.ProjectionPlaneDistance / top.Z())
		bottom = bottom.Mul(camera.ProjectionPlaneDistance / bottom.Z())

		projectionPlaneTopLeft = mgl32.Vec3{left.X(), top.Y(), camera.ProjectionPlaneDistance}
		projectionPlaneBottomRight = mgl32.Vec3{right.X(), bottom.Y(), camera.ProjectionPlaneDistance}
	} else {
		ortographicHalfWidth := camera.OrtographicSize * (float32(resolutionWidth) / float32(resolutionHeight))
		projectionPlaneTopLeft = mgl32.Vec3{-ortographicHalfWidth, camera.OrtographicSize, camera.ProjectionPlaneDistance}
		projectionPlaneBottomRight = mgl32.Vec3{ortographicHalfWidth, -camera.OrtographicSize, camera.ProjectionPlaneDistance}
	}

	verticalStep := (projectionPlaneTopLeft.Y() - projectionPlaneBottomRight.Y()) / float32(resolutionHeight)
	horizontalStep := (projectionPlaneBottomRight.X() - projectionPlaneTopLeft.X()) / float32(resolutionWidth)

	for j := 0; j < resolutionHeight; j++ {
		for i := 0; i < resolutionWidth; i++ {
			var originCameraSpace mgl32.Vec3
			var dir mgl32.Vec3

			x := projectionPlaneTopLeft.X() + horizontalStep*float32(i)
			y := projectionPlaneTopLeft.Y() - verticalStep*float32(j)

			originCameraSpace = mgl32.Vec3{x, y, -camera.ProjectionPlaneDistance}
			origin := mgl32.TransformCoordinate(originCameraSpace, camera.Transform)

			if camera.Projection == Perspective {
				// Camera local origin is always at 0,0,0 so the normalized
				// ray origin is it's direction
				dir = origin.Sub(camera.Transform.Col(3).Vec3()).Normalize()
			} else {
				dir = mgl32.TransformCoordinate(mgl32.Vec3{0, 0, -1}, camera.Transform).Sub(camera.Transform.Col(3).Vec3()).Normalize()
			}

			ray := Ray{
				X:         i,
				Y:         j,
				Bounce:    0,
				Origin:    origin,
				Direction: dir,
			}

			rays[j*resolutionWidth+i] = ray
		}
	}

	return rays
}
