package models

import "image/color"

type Material struct {
	Albedo   color.Color
	Specular float32
}
