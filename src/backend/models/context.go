package models

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

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
	BounceLimit   uint8
	BounceRays    int
	WorkerID      int
}

type RenderPass struct {
	TaskID  int
	RNGSeed int64
	XOffset int
	YOffset int
	Width   int
	Height  int
}

func (context *RenderContext) Initialize() error {
	if math.Abs(float64(context.Camera.Transform.Trace())) < 0.001 {
		context.Camera.Transform = mgl32.Ident4()
	}

	if context.Width < 0 {
		context.Width = 0
	}
	if context.Height < 0 {
		context.Height = 0
	}

	context.Scene.LinkMaterials()
	if len(context.ObjBuffer) > 0 && len(context.MtlBuffer) > 0 {
		options := &gwob.ObjParserOptions{LogStats: true, Logger: func(msg string) { println(msg) }}
		obj, err := gwob.NewObjFromBuf("scene", []byte(context.ObjBuffer), options)
		if err != nil {
			return err
		}

		mtl, err := gwob.ReadMaterialLibFromBuf([]byte(context.MtlBuffer), nil)
		if err != nil {
			return err
		}

		context.Object = obj
		context.MaterialLib = &mtl

		context.DebugMaterial = &gwob.Material{
			Name:  "Debug",
			Kd:    [3]float32{1, 0, 1},
			Ka:    [3]float32{1, 1, 1},
			Ks:    [3]float32{0.5, 0.5, 0.5},
			Ni:    1.0,
			D:     1.0,
			Illum: 2,
			Tr:    0.0,
		}

		context.Triangles = make([]Triangle, 0)

		for _, group := range context.Object.Groups {
			// Each group is an independent object
			material, found := context.MaterialLib.Lib[group.Usemtl]
			if !found {
				material = context.DebugMaterial
			}

			for index := group.IndexBegin; index < group.IndexBegin+group.IndexCount; index += 3 {
				strideIndex0 := context.Object.Indices[index]
				strideIndex1 := context.Object.Indices[index+1]
				strideIndex2 := context.Object.Indices[index+2]
				c0, c1, c2 := context.Object.VertexCoordinates(strideIndex0)
				c3, c4, c5 := context.Object.VertexCoordinates(strideIndex1)
				c6, c7, c8 := context.Object.VertexCoordinates(strideIndex2)

				v0 := mgl32.Vec3{c0, c1, c2}
				v1 := mgl32.Vec3{c3, c4, c5}
				v2 := mgl32.Vec3{c6, c7, c8}

				tri := Triangle{
					Vertices: [3]mgl32.Vec3{v0, v1, v2},
					Normal:   v1.Sub(v0).Cross(v2.Sub(v0)).Normalize(),
					Material: material,
					Edge0:    v1.Sub(v0),
					Edge1:    v2.Sub(v1),
					Edge2:    v0.Sub(v2),
				}
				context.Triangles = append(context.Triangles, tri)
			}
		}
	}
	return nil
}
