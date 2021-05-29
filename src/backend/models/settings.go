package models

import "github.com/go-gl/mathgl/mgl32"

type RenderSettings struct {
	DrawSurfaceNormal   bool
	GammaCorrection     bool
	Gamma               float32
	LightSampleRays     int
	BounceLimit         uint8
	LightIntensity      float32
	DebugLightSize      float32
	ForceDebugLight     bool
	DebugLightAtCamera  bool
	DebugLightTransform mgl32.Mat4
}
