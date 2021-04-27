package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
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

	context, err := parseRenderContext(args[0].String())
	if err != nil {
		return handleError(err, &result)
	}

	//println("Rendering context:")
	//fmt.Printf("%#v \n", context)

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, context.Width, context.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Spawn initial rays
	rays := context.Camera.SpawnRays(context.Width, context.Height)

	// Trace
	for _, ray := range rays {
		color := process.Trace(context, &ray)
		result.ImageData.Set(ray.X, ray.Y, color)
	}

	return result.Output()
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

	if context.Camera.Transform.Trace() == 0 {
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
		obj, err := gwob.NewObjFromBuf("scene", []byte(context.ObjBuffer), nil)
		if err != nil {
			return context, err
		}

		mtl, err := gwob.ReadMaterialLibFromBuf([]byte(context.MtlBuffer), nil)
		if err != nil {
			return context, err
		}

		context.Obj = obj
		context.Mtl = &mtl
	}

	return context, err
}
