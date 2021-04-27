package models

import "github.com/udhos/gwob"

type RenderContext struct {
	Width         int
	Height        int
	Camera        Camera
	Scene         Scene
	Settings      RenderSettings
	ObjBuffer     string
	MtlBuffer     string
	Object        *gwob.Obj
	MaterialLib   *gwob.MaterialLib
	DebugMaterial *gwob.Material
	Triangles     []Triangle
}
