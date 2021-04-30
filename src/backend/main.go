package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"raytracer/models"
	"raytracer/process"
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

	progressUpdate(0.0, "RenderContext.Initialize")

	context.Initialize()

	progressUpdate(1.0, "RenderContext.Initialize")
	return nil
}

// Renders a region given by the parameters
func render(this js.Value, args []js.Value) interface{} {
	println("Go WebAssembly render call")

	result := models.RenderResult{}

	pass, err := parseRenderPass(args[0].String())
	if err != nil {
		return handleError(err, &result)
	}

	progressUpdate(0.0, "spawnRays")

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, pass.Width, pass.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Spawn initial rays
	rays := context.Camera.SpawnRays(pass.XOffset, pass.YOffset, pass.Width, pass.Height)

	progressUpdate(1.0, "spawnRays")

	progressUpdate(0.0, "trace")
	updateInterval := int(float32(len(rays)) / 100.0)
	updateIndex := 0

	println("Initial rays", len(rays))
	// Trace
	for i, ray := range rays {
		if i > updateIndex+updateInterval {
			updateIndex = i
			progress := float32(updateIndex) / float32(len(rays))
			progressUpdate(progress, "trace")
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
		WorkerID int     `json:"workerId"`
	}{
		progress,
		event,
		context.WorkerID,
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
