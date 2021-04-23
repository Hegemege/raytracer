package models

import (
	"encoding/json"
	"image"
)

type RenderResult struct {
	ExitCode  int         `json:"exitCode"`
	Message   string      `json:"message"`
	ImageData *image.RGBA `json:"imageData"`
}

func (res *RenderResult) Output() string {
	data, err := json.Marshal(&res)
	if err != nil {
		panic(err)
	}

	return string(data)
}
