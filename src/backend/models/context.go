package models

import "github.com/udhos/gwob"

type RenderContext struct {
	Width     int
	Height    int
	Camera    Camera
	Scene     Scene
	Settings  RenderSettings
	ObjBuffer string
	MtlBuffer string
	Obj       *gwob.Obj
	Mtl       *gwob.MaterialLib
}
