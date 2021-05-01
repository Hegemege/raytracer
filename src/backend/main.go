package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"raytracer/models"
	"raytracer/process"
	"raytracer/utility"
	"syscall/js"
)

// Globals kept by the same WebWorker for multiple calls
var context *models.RenderContext

func main() {
	println("Go WebAssembly main")

	js.Global().Set("render", js.FuncOf(render))
	js.Global().Set("initialize", js.FuncOf(initialize))

	<-make(chan bool)
}

// Initialize is executed by all webworkers to initialize the rendering context
func initialize(this js.Value, args []js.Value) interface{} {
	println("Go WebAssembly initialize")

	ctx, err := readRenderContext(args[0].String())
	if err != nil {
		panic(err)
	}
	context = ctx

	utility.ProgressUpdate(0.0, "RenderContext.Initialize", -1)

	context.Initialize()

	utility.ProgressUpdate(1.0, "RenderContext.Initialize", -1)
	return nil
}

// Renders a region given by the parameters
func render(this js.Value, args []js.Value) interface{} {
	if context.Settings.Debug {
		println("Go WebAssembly render call")
	}

	result := models.RenderResult{}

	pass, err := parseRenderPass(args[0].String())
	if err != nil {
		return handleError(err, &result)
	}

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, pass.Width, pass.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Spawn initial rays
	rays := context.Camera.SpawnRays(pass.XOffset, pass.YOffset, pass.Width, pass.Height, context.Width, context.Height, pass.TaskID)

	utility.ProgressUpdate(0.0, "trace", pass.TaskID)
	updateInterval := int(float32(len(rays)) / 10.0)
	updateIndex := 0

	// Trace
	for i, ray := range rays {
		if i > updateIndex+updateInterval {
			updateIndex = i
			progress := float32(updateIndex) / float32(len(rays))
			utility.ProgressUpdate(progress, "trace", pass.TaskID)
		}
		rayColor := process.Trace(context, &ray)
		r, g, b, _ := result.ImageData.At(ray.X, ray.Y).RGBA()
		r += uint32(float32(rayColor.R) / float32(context.Camera.RaysPerPixel))
		g += uint32(float32(rayColor.G) / float32(context.Camera.RaysPerPixel))
		b += uint32(float32(rayColor.B) / float32(context.Camera.RaysPerPixel))

		result.ImageData.Set(ray.X, ray.Y, color.RGBA{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
			A: 255,
		})
	}
	utility.ProgressUpdate(1.0, "trace", pass.TaskID)

	utility.ProgressUpdate(0.0, "output", pass.TaskID)
	output := result.Output()
	utility.ProgressUpdate(1.0, "output", pass.TaskID)

	return output
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

func readRenderContext(rawContext string) (*models.RenderContext, error) {
	context := &models.RenderContext{}
	err := json.Unmarshal([]byte(rawContext), context)
	if err != nil {
		return nil, err
	}
	return context, nil
}

func parseRenderPass(raw string) (*models.RenderPass, error) {
	pass := &models.RenderPass{}
	err := json.Unmarshal([]byte(raw), pass)
	if err != nil {
		return nil, err
	}
	rand.Seed(pass.RNGSeed)
	return pass, nil
}
