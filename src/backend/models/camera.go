package models

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distmv"
	"gonum.org/v1/gonum/stat/samplemv"
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

	projectionPlaneTopLeft mgl32.Vec3
	horizontalStep         float32
	verticalStep           float32

	sampler    *samplemv.Halton
	batch      *mat.Dense
	index      int
	maxSamples int
}

func (camera *Camera) Initialize(totalWidth int, totalHeight int) {
	// Create sampler
	camera.maxSamples = 12345
	camera.batch = mat.NewDense(camera.maxSamples, 2, nil)
	camera.sampler = &samplemv.Halton{
		Kind: samplemv.Owen,
		Q:    distmv.NewUnitUniform(2, nil),
	}

	camera.sampler.Sample(camera.batch)

	var projectionPlaneTopLeft mgl32.Vec3
	var projectionPlaneBottomRight mgl32.Vec3

	if camera.Projection == Perspective {
		verticalHalfAngle := math.Pi * (camera.FieldOfView / 2.0) / 180.0
		horizontalHalfAngle := verticalHalfAngle * (float32(totalWidth) / float32(totalHeight))

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
		ortographicHalfWidth := camera.OrtographicSize * (float32(totalWidth) / float32(totalHeight))
		projectionPlaneTopLeft = mgl32.Vec3{-ortographicHalfWidth, camera.OrtographicSize, camera.ProjectionPlaneDistance}
		projectionPlaneBottomRight = mgl32.Vec3{ortographicHalfWidth, -camera.OrtographicSize, camera.ProjectionPlaneDistance}
	}

	camera.verticalStep = (projectionPlaneTopLeft.Y() - projectionPlaneBottomRight.Y()) / float32(totalHeight)
	camera.horizontalStep = (projectionPlaneBottomRight.X() - projectionPlaneTopLeft.X()) / float32(totalWidth)
	camera.projectionPlaneTopLeft = projectionPlaneTopLeft
}

func (camera *Camera) samplePixel() mgl32.Vec2 {
	// Get Halton sample
	sample := mgl32.Vec2{
		float32(camera.batch.At(camera.index, 0)),
		float32(camera.batch.At(camera.index, 1)),
	}

	camera.index = (camera.index + 1) % camera.maxSamples

	return sample
}

func (camera *Camera) GetCameraRay(xoffset int, yoffset int, x int, y int) *Ray {

	var dir mgl32.Vec3

	/*
		rx := rand.Float32()*0.5 - 1
		ry := rand.Float32()*0.5 - 1

		lx := camera.projectionPlaneTopLeft.X() + camera.horizontalStep*(float32(xoffset+x)+rx)
		ly := camera.projectionPlaneTopLeft.Y() - camera.verticalStep*(float32(yoffset+y)+ry)
	*/

	sample := camera.samplePixel()
	lx := camera.projectionPlaneTopLeft.X() + camera.horizontalStep*(float32(xoffset+x)+sample.X())
	ly := camera.projectionPlaneTopLeft.Y() - camera.verticalStep*(float32(yoffset+y)+sample.Y())

	originCameraSpace := mgl32.Vec3{lx, ly, -camera.ProjectionPlaneDistance}
	origin := mgl32.TransformCoordinate(originCameraSpace, camera.Transform)

	if camera.Projection == Perspective {
		// Camera local origin is always at 0,0,0 so the normalized
		// ray origin is it's direction
		dir = origin.Sub(camera.Transform.Col(3).Vec3()).Normalize()
	} else {
		dir = mgl32.TransformCoordinate(mgl32.Vec3{0, 0, -1}, camera.Transform).Sub(camera.Transform.Col(3).Vec3()).Normalize()
	}

	ray := NewRay(origin, dir, 0, x-xoffset, y-yoffset)

	return ray
}
