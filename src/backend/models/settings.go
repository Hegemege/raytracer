package models

type RenderSettings struct {
	DrawSurfaceNormal bool
	GammaCorrection   bool
	Gamma             float32
	LightSampleRays   int
	BounceLimit       uint8
	LightIntensity    float32
	DebugLightSize    float32
}
