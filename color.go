package materials

import (
	"fmt"
	"image/color"
)

// AlphaMultiply returns the input color with the new alpha value.
// The Red, Green, and Blue components are automatically premultiplied
// for the new alpha.
func AlphaMultiply(in color.RGBA, newAlpha uint8) color.RGBA {
	in.R = uint8(int(in.R) * int(newAlpha) / 255)
	in.G = uint8(int(in.G) * int(newAlpha) / 255)
	in.B = uint8(int(in.B) * int(newAlpha) / 255)
	in.A = uint8(int(in.A) * int(newAlpha) / 255)
	return in
}

func ColorString(in color.RGBA) string {
	return fmt.Sprintf("%02x%02x%02x%02x", in.R, in.G, in.B, in.A)
}
