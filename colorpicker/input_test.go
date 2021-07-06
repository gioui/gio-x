package colorpicker

import (
	"encoding/hex"
	"gioui.org/layout"
	"image/color"
	"testing"
)

func TestInputSetColor(t *testing.T) {
	col := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	for _, in := range []ColorInput{
		NewPicker(nil),
		NewAlphaSlider(),
		NewHexEditor(nil),
		NewRgbEditor(nil),
		NewHsvEditor(nil),
		NewToggle(nil, NewHexEditor(nil)),
		NewColorSelection(nil, layout.SE),
		NewMux(),
	} {
		if in.Changed() {
			t.Fail()
		}
		in.SetColor(col)
		if in.Changed() {
			t.Fail()
		}
		if in.Color() != col {
			t.Fail()
		}
	}
}

func TestMux(t *testing.T) {
	col := color.NRGBA{R: 134, G: 99, B: 12, A: 0}
	hexe := NewHexEditor(nil)
	rgb := NewRgbEditor(nil)
	alpha := NewAlphaSlider()
	mux := NewMux(hexe, rgb, alpha)
	hexe.hex.editor.SetText(hex.EncodeToString([]byte{col.R, col.G, col.B}))
	if !mux.Changed() {
		t.Fail()
	}
	if mux.Color() != col {
		t.Fail()
	}
	if rgb.Color() != col {
		t.Fail()
	}
	if alpha.Color() != col {
		t.Fail()
	}
}
