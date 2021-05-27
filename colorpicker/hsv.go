// Implementation of HSV colorspace

package colorpicker

import (
	"image/color"
)

// HSV is a cylindrical rgb model that remaps the RGB primary colors
// into dimensions that are easier for humans to understand.
type HSVColor struct {
	// H stands for hue and is the rgb portion of the model,
	// expressed as a number from 0 to 360 degrees:
	// Red falls between 0 and 60 degrees.
	// Yellow falls between 61 and 120 degrees.
	// Green falls between 121 and 180 degrees.
	// Cyan falls between 181 and 240 degrees.
	// Blue falls between 241 and 300 degrees.
	// Magenta falls between 301 and 360 degrees.
	H float32
	// S stands for saturation and describes the amount of gray in a particular rgb,from 0 to 1.
	// Reducing this component toward zero introduces more gray and produces a faded effect.
	S float32
	// V stands for value and works in conjunction with saturation and describes the brightness or intensity of the rgb,
	// from 0 to 1, where 0 is completely black, and 1 is the brightest and reveals the most rgb.
	V float32
}

func HsvToRgb(hsv HSVColor) color.NRGBA {
	v := uint8(hsv.V * 255)
	if hsv.S == 0.0 {
		return color.NRGBA{R: v, G: v, B: v, A: 0xff}
	}

	hh := hsv.H
	if hh >= 360.0 {
		hh = 0.0
	}

	hh /= 60.0
	i := int(hh)
	ff := hh - float32(i)
	p := uint8(hsv.V * (1 - hsv.S) * 255)
	q := uint8(hsv.V * (1 - (hsv.S * ff)) * 255)
	t := uint8(hsv.V * (1 - (hsv.S * (1 - ff))) * 255)

	rgb := color.NRGBA{A: 0xff}
	switch i {
	case 0:
		rgb.R = v
		rgb.G = t
		rgb.B = p
	case 1:
		rgb.R = q
		rgb.G = v
		rgb.B = p
	case 2:
		rgb.R = p
		rgb.G = v
		rgb.B = t
	case 3:
		rgb.R = p
		rgb.G = q
		rgb.B = v
	case 4:
		rgb.R = t
		rgb.G = p
		rgb.B = v
	case 5:
		rgb.R = v
		rgb.G = p
		rgb.B = q
	}
	return rgb
}

func RgbToHsv(rgb color.NRGBA) HSVColor {
	var hsv HSVColor

	rgbMin := float32(min(min(rgb.R, rgb.G), rgb.B))
	rgbMax := float32(max(max(rgb.R, rgb.G), rgb.B))

	hsv.V = rgbMax / 255
	delta := rgbMax - rgbMin
	if delta == 0 {
		return hsv
	}
	if rgbMax == 0 {
		//hsv.H = NaN
		return hsv
	} else {
		hsv.S = delta / rgbMax // s
	}
	r, g, b := float32(rgb.R), float32(rgb.G), float32(rgb.B)
	if r == rgbMax {
		hsv.H = (g - b) / delta
	} else if g == rgbMax {
		hsv.H = 2.0 + (b-r)/delta
	} else {
		hsv.H = 4.0 + (r-g)/delta
	}
	hsv.H *= 60
	if hsv.H < 0 {
		hsv.H += 360
	}
	return hsv
}

func min(a, b uint8) uint8 {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b uint8) uint8 {
	if a > b {
		return a
	} else {
		return b
	}
}
