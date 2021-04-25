package models

type RenderContext struct {
	Width          int
	Height         int
	CameraSettings Camera
	Scene          Scene
}
