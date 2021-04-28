package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"math"
	"raytracer/models"
	"raytracer/process"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

func main() {
	println("Go WebAssembly initialized")

	js.Global().Set("render", js.FuncOf(render))

	<-make(chan bool)
}

func render(this js.Value, args []js.Value) interface{} {
	println("Go WebAssembly render call")

	result := models.RenderResult{}

	progressUpdate(0.0, "parseRenderContext")

	context, err := parseRenderContext(args[0].String())
	if err != nil {
		return handleError(err, &result)
	}

	progressUpdate(1.0, "parseRenderContext")

	progressUpdate(0.0, "spawnRays")

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, context.Width, context.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Spawn initial rays
	rays := context.Camera.SpawnRays(context.Width, context.Height)

	progressUpdate(1.0, "spawnRays")

	progressUpdate(0.0, "trace")
	updateInterval := int(float32(len(rays)) / 100.0)
	updateIndex := 0
	// Trace
	for i, ray := range rays {
		if i > updateIndex+updateInterval {
			updateIndex = i
			progress := float32(updateIndex) / float32(len(rays))
			progressUpdate(progress, "trace")
		}

		color := process.Trace(context, &ray)
		result.ImageData.Set(ray.X, ray.Y, color)
	}
	progressUpdate(1.0, "trace")

	progressUpdate(0.0, "output")
	output := result.Output()
	progressUpdate(1.0, "output")

	return output
}

func progressUpdate(progress float32, event string) {
	data := struct {
		Progress float32 `json:"progress"`
		Event    string  `json:"event"`
	}{
		progress, event,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	js.Global().Call("progressUpdate", string(raw))
}

func handleError(err error, result *models.RenderResult) string {
	if result == nil {
		result = &models.RenderResult{}
	}

	result.ExitCode = -1
	if err == nil {
		result.Message = "Unknown error"
	} else {
		result.Message = err.Error()
	}

	result.ImageData = image.NewRGBA(image.Rect(0, 0, 500, 500))

	println("ERROR:", result.Message)

	return result.Output()
}

func parseRenderContext(rawContext string) (*models.RenderContext, error) {
	context := &models.RenderContext{}
	err := json.Unmarshal([]byte(rawContext), context)
	if err != nil {
		return context, err
	}

	context.BounceLimit = 1
	context.BounceRays = 100

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
			return context, err
		}

		mtl, err := gwob.ReadMaterialLibFromBuf([]byte(context.MtlBuffer), nil)
		if err != nil {
			return context, err
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

		context.Triangles = make([]models.Triangle, 0)

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

				tri := models.Triangle{
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

	return context, err
}
