package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"raytracer/models"
	"raytracer/process"
	"raytracer/utility"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
)

// Globals kept by the same WebWorker for multiple calls
var context *models.RenderContext
var activeRenderKey int = -1

func main() {
	println("Go WebAssembly main")

	js.Global().Set("initialize", js.FuncOf(initialize))
	js.Global().Set("buildBVH", js.FuncOf(buildBVH))
	js.Global().Set("loadBVH", js.FuncOf(loadBVH))
	js.Global().Set("render", js.FuncOf(render))
	js.Global().Set("incrementalRender", js.FuncOf(incrementalRender))
	js.Global().Set("initializeIncrementalRender", js.FuncOf(initializeIncrementalRender))

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

	rawTextureData := make([]*[]uint8, 0)
	// Read texture data from the rest of the inputs
	for i := 1; i < len(args); i++ {
		imageArr := args[i]
		inBuf := make([]uint8, imageArr.Get("byteLength").Int())
		js.CopyBytesToGo(inBuf, imageArr)
		rawTextureData = append(rawTextureData, &inBuf)
	}

	utility.ProgressUpdate(0.0, "RenderContext.Initialize", -1, 0)
	context.Initialize(rawTextureData)
	utility.ProgressUpdate(1.0, "RenderContext.Initialize", -1, 0)
	return nil
}

func buildBVH(this js.Value, args []js.Value) interface{} {
	utility.ProgressUpdate(0.0, "RenderContext.BuildBVH", -1, 0)
	bvh := context.BuildBVH()
	utility.ProgressUpdate(1.0, "RenderContext.BuildBVH", -1, 0)

	rawBVH, err := json.Marshal(bvh)
	if err != nil {
		panic(err)
	}

	return string(rawBVH)
}

func loadBVH(this js.Value, args []js.Value) interface{} {
	bvh := &models.BVH{}
	err := json.Unmarshal([]byte(args[0].String()), bvh)
	if err != nil {
		panic(err)
	}

	utility.ProgressUpdate(0.0, "RenderContext.LoadBVH", -1, 0)
	context.LoadBVH(bvh)
	utility.ProgressUpdate(1.0, "RenderContext.LoadBVH", -1, 0)

	return nil
}

// Renders a region given by the parameters
func render(this js.Value, args []js.Value) interface{} {
	if context.Debug {
		println("Go WebAssembly render call")
	}

	result := models.RenderResult{}

	pass, err := parseRenderPass(args[0].String())
	if err != nil {
		return handleError(err, &result)
	}

	pass.Initialize(context)

	if activeRenderKey != pass.RenderKey {
		activeRenderKey = pass.RenderKey
		context.Rays = 0
	}
	pass.Camera.Initialize(pass.TotalWidth, pass.TotalHeight)

	// Fill with black
	result.ImageData = image.NewRGBA(image.Rect(0, 0, pass.Width, pass.Height))
	draw.Draw(result.ImageData, result.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	pixelCount := pass.Width * pass.Height
	rayCount := pixelCount * pass.Camera.RaysPerPixel

	utility.ProgressUpdate(0.0, "trace", pass.TaskID, context.Rays)
	updateInterval := int(float32(rayCount) / 10.0)
	updateIndex := 0

	colors := make([]mgl32.Vec3, pass.Width*pass.Height)
	for j := 0; j < pass.Height; j++ {
		for i := 0; i < pass.Width; i++ {
			colors[i+j*pass.Width] = mgl32.Vec3{0, 0, 0}
		}
	}

	// Trace
	ri := 0
	for i := 0; i < pixelCount; i++ {
		x := i % pass.Width
		y := i / pass.Width
		pixelColor := mgl32.Vec3{0, 0, 0}
		for j := 0; j < pass.Camera.RaysPerPixel; j++ {
			if ri > updateIndex+updateInterval {
				updateIndex = ri
				progress := float32(updateIndex) / float32(rayCount)
				utility.ProgressUpdate(progress, "trace", pass.TaskID, context.Rays)
			}
			ri += 1

			ray := pass.Camera.GetCameraRay(pass.XOffset, pass.YOffset, x, y)

			rayColor := process.Trace(context, pass, ray)
			pixelColor = pixelColor.Add(rayColor)

		}
		colors[i] = pixelColor
	}

	utility.ProgressUpdate(1.0, "trace", pass.TaskID, context.Rays)

	utility.ProgressUpdate(0.0, "output", pass.TaskID, context.Rays)

	for j := 0; j < pass.Height; j++ {
		for i := 0; i < pass.Width; i++ {
			c := colors[i+j*pass.Width]
			c = c.Mul(1.0 / float32(pass.Camera.RaysPerPixel))

			// Gamma correction
			if pass.Settings.GammaCorrection {
				gamma := float64(1.0 / pass.Settings.Gamma)
				c = mgl32.Vec3{
					float32(math.Pow(float64(c.X()), gamma)),
					float32(math.Pow(float64(c.Y()), gamma)),
					float32(math.Pow(float64(c.Z()), gamma)),
				}
			}

			c = utility.ClampColor(c)

			result.ImageData.SetRGBA(i, j, color.RGBA{
				R: uint8(255 * c.X()),
				G: uint8(255 * c.Y()),
				B: uint8(255 * c.Z()),
				A: 255,
			})
		}
	}

	output := result.Output()
	utility.ProgressUpdate(1.0, "output", pass.TaskID, context.Rays)

	return output
}

var incrementalRenderPass *models.RenderPass
var incrementalResult models.RenderResult
var incrementalRenderingIndex int
var incrementalRenderReportIndex int
var incrementalRenderColors []mgl32.Vec3

func initializeIncrementalRender(this js.Value, args []js.Value) interface{} {
	if context.Debug {
		println("Go WebAssembly initializeIncrementalRender call")
	}
	incrementalResult = models.RenderResult{}
	incrementalRenderingIndex = 0
	incrementalRenderReportIndex = 0

	pass, err := parseRenderPass(args[0].String())
	if err != nil {
		return handleError(err, &incrementalResult)
	}

	incrementalRenderPass = pass

	incrementalRenderPass.Initialize(context)

	if activeRenderKey != incrementalRenderPass.RenderKey {
		activeRenderKey = incrementalRenderPass.RenderKey
		context.Rays = 0
	}
	incrementalRenderPass.Camera.Initialize(incrementalRenderPass.TotalWidth, incrementalRenderPass.TotalHeight)

	// Fill with black
	incrementalResult.ImageData = image.NewRGBA(image.Rect(0, 0, incrementalRenderPass.Width, incrementalRenderPass.Height))
	draw.Draw(incrementalResult.ImageData, incrementalResult.ImageData.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	incrementalRenderColors = make([]mgl32.Vec3, incrementalRenderPass.Width*incrementalRenderPass.Height)
	for j := 0; j < incrementalRenderPass.Height; j++ {
		for i := 0; i < incrementalRenderPass.Width; i++ {
			incrementalRenderColors[i+j*incrementalRenderPass.Width] = mgl32.Vec3{0, 0, 0}
		}
	}

	return nil
}

// Renders a region given by the parameters
func incrementalRender(this js.Value, args []js.Value) interface{} {
	if context.Debug {
		println("Go WebAssembly incrementalRender call")
	}

	pixelCount := incrementalRenderPass.Width * incrementalRenderPass.Height

	// Send first 0% progress
	if incrementalRenderingIndex == 0 {
		utility.ProgressUpdate(0.0, "trace", incrementalRenderPass.TaskID, context.Rays)
	}

	totalPixelCount := pixelCount * incrementalRenderPass.Camera.RaysPerPixel

	updateInterval := int(float32(totalPixelCount) / 10.0)

	// Trace
	ri := incrementalRenderingIndex * pixelCount
	for i := 0; i < pixelCount; i++ {
		x := i % incrementalRenderPass.Width
		y := i / incrementalRenderPass.Width

		if ri > incrementalRenderReportIndex+updateInterval {
			incrementalRenderReportIndex = ri
			progress := float32(incrementalRenderReportIndex) / float32(totalPixelCount)
			utility.ProgressUpdate(progress, "trace", incrementalRenderPass.TaskID, context.Rays)
		}
		ri += 1

		ray := incrementalRenderPass.Camera.GetCameraRay(incrementalRenderPass.XOffset, incrementalRenderPass.YOffset, x, y)

		rayColor := process.Trace(context, incrementalRenderPass, ray)

		incrementalRenderColors[i] = incrementalRenderColors[i].Add(rayColor)
	}

	incrementalRenderingIndex += 1

	// Send last 100% progress
	if incrementalRenderingIndex == incrementalRenderPass.Camera.RaysPerPixel {
		utility.ProgressUpdate(1.0, "trace", incrementalRenderPass.TaskID, context.Rays)
	}

	for j := 0; j < incrementalRenderPass.Height; j++ {
		for i := 0; i < incrementalRenderPass.Width; i++ {
			c := incrementalRenderColors[i+j*incrementalRenderPass.Width]
			c = c.Mul(1.0 / float32(incrementalRenderingIndex))

			// Gamma correction
			if incrementalRenderPass.Settings.GammaCorrection {
				gamma := float64(1.0 / incrementalRenderPass.Settings.Gamma)
				c = mgl32.Vec3{
					float32(math.Pow(float64(c.X()), gamma)),
					float32(math.Pow(float64(c.Y()), gamma)),
					float32(math.Pow(float64(c.Z()), gamma)),
				}
			}

			c = utility.ClampColor(c)

			incrementalResult.ImageData.SetRGBA(i, j, color.RGBA{
				R: uint8(255 * c.X()),
				G: uint8(255 * c.Y()),
				B: uint8(255 * c.Z()),
				A: 255,
			})
		}
	}

	output := incrementalResult.Output()

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
