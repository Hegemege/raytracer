package models

type RenderContext struct {
	Width    int
	Height   int
	Camera   Camera
	Scene    Scene
	Settings RenderSettings
}
