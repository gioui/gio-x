package materials

import (
	"image/color"
)

// AlphaMultiply returns the input color with the new alpha value.
// The Red, Green, and Blue components are automatically premultiplied
// for the new alpha.
func AlphaMultiply(c color.RGBA, a uint8) color.RGBA {
	return color.RGBA{
		R: uint8(int(c.R) * int(a) / 255),
		G: uint8(int(c.G) * int(a) / 255),
		B: uint8(int(c.B) * int(a) / 255),
		A: a,
	}
}

// AlphaPalette is the set of alpha values to be applied for certain
// material design states like hover, selected, etc...
type AlphaPalette struct {
	Hover, Selected uint8
}
