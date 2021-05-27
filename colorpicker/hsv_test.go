package colorpicker

import (
	"fmt"
	"image/color"
	"math"
	"testing"
)

func TestRgbHsvConversion(t *testing.T) {
	for _, v := range []struct {
		rgb color.NRGBA
		hsv HSVColor
	}{
		{color.NRGBA{A: 0xff}, HSVColor{H: 0, S: 0, V: 0}},                                    // black
		{color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, HSVColor{H: 0, S: 0, V: 1}},         // white
		{color.NRGBA{R: 0xff, A: 0xff}, HSVColor{H: 0, S: 1, V: 1}},                           // red
		{color.NRGBA{G: 0xff, A: 0xff}, HSVColor{H: 120, S: 1, V: 1}},                         // lime
		{color.NRGBA{B: 0xff, A: 0xff}, HSVColor{H: 240, S: 1, V: 1}},                         // blue
		{color.NRGBA{R: 0xff, G: 0xff, A: 0xff}, HSVColor{H: 60, S: 1, V: 1}},                 // yellow
		{color.NRGBA{G: 0xff, B: 0xff, A: 0xff}, HSVColor{H: 180, S: 1, V: 1}},                // cyan
		{color.NRGBA{R: 0xff, B: 0xff, A: 0xff}, HSVColor{H: 300, S: 1, V: 1}},                // magenta
		{color.NRGBA{R: 0xbf, G: 0xbf, B: 0xbf, A: 0xff}, HSVColor{H: 0, S: 0, V: .75}},       // silver
		{color.NRGBA{R: 0x7f, G: 0x7f, B: 0x7f, A: 0xff}, HSVColor{H: 0, S: 0, V: .5}},        // gray
		{color.NRGBA{R: 0x7f, A: 0xff}, HSVColor{H: 0, S: 1, V: .5}},                          // maroon
		{color.NRGBA{R: 0xff, G: 0x7f, A: 0xff}, HSVColor{H: 60, S: 1, V: .5}},                // olive
		{color.NRGBA{G: 0x7f, A: 0xff}, HSVColor{H: 120, S: 1, V: .5}},                        // green
		{color.NRGBA{R: 0x7f, B: 0x7f, A: 0xff}, HSVColor{H: 300, S: 1, V: .5}},               // purple
		{color.NRGBA{G: 0x7f, B: 0x7f, A: 0xff}, HSVColor{H: 180, S: 1, V: .5}},               // teal
		{color.NRGBA{B: 0x7f, A: 0xff}, HSVColor{H: 240, S: 1, V: .5}},                        // navy
		{color.NRGBA{R: 0xfd, G: 0x01, B: 0xc7, A: 0xff}, HSVColor{H: 313, S: .996, V: .992}}, // pink
		{color.NRGBA{R: 0xff, G: 0x7f, B: 0x00, A: 0xff}, HSVColor{H: 30, S: 1, V: 1}},        // orange
	} {
		testHsvToRgb(v.hsv, RgbToHsv(v.rgb))
		testRgbToHsv(v.rgb, HsvToRgb(v.hsv))
	}
}

func testRgbToHsv(a color.NRGBA, b color.NRGBA) {
	if a.R != b.R && a.G != b.G && a.B != b.B {
		panic(fmt.Sprintf("not equal: %v, %v", a, b))
	}
}

func testHsvToRgb(a HSVColor, b HSVColor) {
	aH := math.Round(float64(a.H))
	bH := math.Round(float64(b.H))
	aV := math.Round(float64(a.V) * 1000)
	bV := math.Round(float64(a.V) * 1000)
	if aH != bH && a.S != b.S && aV == bV {
		panic(fmt.Sprintf("not equal: %v, %v", a, b))
	}
}
