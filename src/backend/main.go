package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"raytracer/models"
	"raytracer/process"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
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

	println("Rendering context:")
	fmt.Printf("%#v \n", context)

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, context.Width, context.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Spawn initial rays
	rays := context.CameraSettings.SpawnRays(context.Width, context.Height)

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

	return result.Output()
}

func parseRenderContext(rawContext string) (*models.RenderContext, error) {
	context := &models.RenderContext{}
	err := json.Unmarshal([]byte(rawContext), context)

	if context.CameraSettings.Transform.Trace() == 0 {
		context.CameraSettings.Transform = mgl32.Ident4()
	}

	if context.Width < 0 {
		context.Width = 0
	}
	if context.Height < 0 {
		context.Height = 0
	}

	// DEBUG: Add some basic shapes/materials
	material := models.Material{
		Albedo:   color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Specular: 0,
	}

	context.Scene = models.Scene{
		Spheres: []models.Sphere{
			{
				Object: models.Object{
					Position: mgl32.Vec3{0, 0, 5},
					Material: &material,
				},
				Radius: 1,
			},
		},
	}

	return context, err
}
