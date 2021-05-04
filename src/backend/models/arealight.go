package models

import (
	"github.com/go-gl/mathgl/mgl32"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distmv"
	"gonum.org/v1/gonum/stat/samplemv"
)

type AreaLight struct {
	Transform mgl32.Mat4
	Size      mgl32.Vec2
	Emission  mgl32.Vec3
	Normal    mgl32.Vec3

	sampler    *samplemv.Halton
	batch      *mat.Dense
	index      int
	maxSamples int
}

func NewAreaLight(transform mgl32.Mat4, size mgl32.Vec2, emission mgl32.Vec3, normal mgl32.Vec3) *AreaLight {
	light := &AreaLight{
		Transform:  transform,
		Size:       size,
		Emission:   emission,
		Normal:     normal,
		maxSamples: 100001,
	}

	light.batch = mat.NewDense(light.maxSamples, 2, nil)
	light.sampler = &samplemv.Halton{
		Kind: samplemv.Owen,
		Q:    distmv.NewUnitUniform(2, nil),
	}

	light.sampler.Sample(light.batch)

	return light
}

func (light *AreaLight) Sample() (mgl32.Vec3, float32) {
	// Get Halton sample
	sample := mgl32.Vec3{
		float32(light.batch.At(light.index, 0)*2-1) * light.Size.X(),
		float32(light.batch.At(light.index, 1)*2-1) * light.Size.Y(),
		0,
	}

	light.index = (light.index + 1) % light.maxSamples

	worldSample := mgl32.TransformCoordinate(sample, light.Transform)
	pdf := 1.0 / (4.0 * light.Size.X() * light.Size.Y())

	return worldSample, pdf
}
