package models

type RenderSettings struct {
	DrawSurfaceNormal bool
	GammaCorrection   bool
	Gamma             float32
	BounceRays        int
	BounceLimit       uint8
}
