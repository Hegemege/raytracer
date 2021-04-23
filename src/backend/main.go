package main

import (
	"encoding/json"
	"fmt"
	"image"
	"raytracer/models"
	"syscall/js"
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
	fmt.Printf("%#v\n", context)

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

	return context, err
}
