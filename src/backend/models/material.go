package models

import "image/color"

type Material struct {
	ID       int
	Albedo   color.RGBA
	Specular float32
}
