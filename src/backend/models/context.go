package models

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

type RenderContext struct {
	Debug         bool
	Scene         Scene
	ObjBuffer     string
	MtlBuffer     string
	RawTextures   []Texture
	Object        *gwob.Obj
	MaterialLib   *gwob.MaterialLib
	DebugMaterial *gwob.Material
	Triangles     []*Triangle
	Light         *AreaLight
	WorkerID      int

	UseBVH         bool
	BVHMaxLeafSize int

	BVH *BVH

	// Statistics
	Rays uint64

	TextureLookup map[string]*Texture
}

type RenderPass struct {
	Camera      Camera
	TotalWidth  int
	TotalHeight int
	TaskID      int
	RNGSeed     int64
	XOffset     int
	YOffset     int
	Width       int
	Height      int
	Settings    RenderSettings
}

func (context *RenderContext) Initialize(rawTextureData []*[]byte) error {

	// Reset stats (if somehow set by client)
	context.Rays = 0

	context.Scene.LinkMaterials()
	if len(context.ObjBuffer) > 0 && len(context.MtlBuffer) > 0 {
		optionsLogger := &gwob.ObjParserOptions{LogStats: context.Debug, Logger: func(msg string) { println(msg) }}
		//options := &gwob.ObjParserOptions{LogStats: context.Debug, Logger: nil}

		obj, err := gwob.NewObjFromBuf("scene", []byte(context.ObjBuffer), optionsLogger)
		if err != nil {
			return err
		}

		mtl, err := gwob.ReadMaterialLibFromBuf([]byte(context.MtlBuffer), optionsLogger)
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

		// Build texture map
		context.TextureLookup = make(map[string]*Texture)
		for i, texture := range context.RawTextures {
			// Read the texture
			rawData := rawTextureData[i]
			t := NewTexture(texture.Name, rawData)
			context.TextureLookup[texture.Name] = t
		}

		// Build triangles
		// TODO: Preallocate triangle array length
		context.Triangles = make([]*Triangle, 0)

		for _, group := range context.Object.Groups {
			// Each group is an independent object
			material, found := context.MaterialLib.Lib[group.Usemtl]
			if !found {
				material = context.DebugMaterial
			}

			triangleIndex := 0

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

				tri := NewTriangle(v0, v1, v2, material, triangleIndex)
				triangleIndex++

				context.Triangles = append(context.Triangles, tri)
			}
		}

		// Parse the area light
		var minx, miny, minz float32 = math.MaxFloat32, math.MaxFloat32, math.MaxFloat32
		var maxx, maxy, maxz float32 = -math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32
		var normal mgl32.Vec3
		var up mgl32.Vec3
		var shortestSide mgl32.Vec3
		var middleSide mgl32.Vec3
		var found bool
		for _, triangle := range context.Triangles {
			if triangle.Material.Name == "Light" {
				found = true
				normal = triangle.Normal
				// Choose the edge to cross normal with to find up
				// as the edge that is shortest. Up will then
				// point from the middle of the arealight towards
				// the long axis
				shortestSide = triangle.GetShortestEdge()
				middleSide = triangle.GetMiddleEdge()
				up = shortestSide.Cross(triangle.Normal).Normalize()
				for _, vertex := range triangle.Vertices {
					if vertex.X() < minx {
						minx = vertex.X()
					}
					if vertex.Y() < miny {
						miny = vertex.Y()
					}
					if vertex.Z() < minz {
						minz = vertex.Z()
					}
					if vertex.X() > maxx {
						maxx = vertex.X()
					}
					if vertex.Y() > maxy {
						maxy = vertex.Y()
					}
					if vertex.Z() > maxz {
						maxz = vertex.Z()
					}
				}
			}
		}

		if found {
			center := mgl32.Vec3{(minx + maxx) / 2.0, (miny + maxy) / 2.0, (minz + maxz) / 2.0}

			//transform := mgl32.LookAtV(normal, center, up).Inv()
			transform := mgl32.Translate3D(center.X(), center.Y(), center.Z())
			transform = transform.Mul4(mgl32.Mat3FromCols(normal.Cross(up), up, normal).Mat4())

			size := mgl32.Vec2{shortestSide.Len() / 2.0, middleSide.Len() / 2.0}
			emission := mgl32.Vec3{100, 100, 100}

			light := NewAreaLight(transform, size, emission, normal)

			context.Light = light
		} else {
			// Render pass Initialize creates a debug light at the camera position
		}
	}

	return nil
}

func (context *RenderContext) BuildBVH() *BVH {
	return BuildBVH(context)
}

func (context *RenderContext) LoadBVH(bvh *BVH) {
	context.BVH = bvh
	context.BVH.Load(context.Triangles)
}

func (pass *RenderPass) Initialize(context *RenderContext) {
	if pass.TotalWidth < 0 {
		pass.TotalWidth = 0
	}
	if pass.TotalHeight < 0 {
		pass.TotalHeight = 0
	}

	if math.Abs(float64(pass.Camera.Transform.Trace())) < 0.001 {
		pass.Camera.Transform = mgl32.Ident4()
	}

	if context.Light == nil {
		// Create debug light
		transform := pass.Camera.Transform
		normal := mgl32.TransformCoordinate(mgl32.Vec3{0, 0, 1}, transform).Sub(transform.Col(3).Vec3())
		size := mgl32.Vec2{1.0, 1.0}
		emission := mgl32.Vec3{100, 100, 100}
		light := NewAreaLight(transform, size, emission, normal)
		context.Light = light
	}
}
