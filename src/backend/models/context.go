package models

type RenderContext struct {
	Width          int `json:"width"`
	Height         int `json:"height"`
	CameraSettings Camera
}
