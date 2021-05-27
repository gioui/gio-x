package colorpicker

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
)

func NewColorSelection(th *material.Theme, drop layout.Direction, inputs ...ColorInput) *ColorSelection {
	mux := &Mux{inputs: inputs}
	return &ColorSelection{
		Dropdown:     drop,
		CornerRadius: unit.Dp(4),
		Inset: layout.Inset{
			Top: unit.Dp(10), Bottom: unit.Dp(10),
			Left: unit.Dp(12), Right: unit.Dp(12),
		},
		Input:   mux,
		clicker: &widget.Clickable{},
		theme:   th,
	}
}

type ColorSelection struct {
	Dropdown     layout.Direction
	CornerRadius unit.Value
	Inset        layout.Inset
	Input        *Mux
	clicker      *widget.Clickable
	active       bool
	theme        *material.Theme
}

func (cf *ColorSelection) Layout(gtx layout.Context) layout.Dimensions {
	cf.event()
	h := int(2.3 * float32(gtx.Metric.Px(cf.theme.TextSize)))
	w := int(float32(h) * 1.5)
	size := image.Point{X: w, Y: int(h)}
	gtx.Constraints = layout.Exact(size)
	drawCheckerboard(gtx)
	dims1 := widget.Border{Color: lightgrey, CornerRadius: cf.CornerRadius, Width: unit.Dp(1)}.
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.ButtonLayoutStyle{
				Background:   cf.Color(),
				CornerRadius: cf.CornerRadius,
				Button:       cf.clicker,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return cf.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Label(cf.theme, cf.theme.TextSize, " ").Layout(gtx)
				})
			})
		})
	if cf.active {
		gtx.Constraints = layout.Exact(image.Point{X: gtx.Px(unit.Dp(210)), Y: gtx.Px(unit.Dp(200))})
		macro := op.Record(gtx.Ops)
		dims2 := cf.Input.Layout(gtx)
		var offset f32.Point
		switch cf.Dropdown {
		case layout.NE:
			offset = f32.Point{X: float32(dims1.Size.X - dims2.Size.X), Y: float32(-dims2.Size.Y)}
		case layout.SE:
			offset = f32.Point{X: float32(dims1.Size.X - dims2.Size.X), Y: float32(dims1.Size.Y)}
		case layout.SW:
			offset = f32.Point{Y: float32(dims1.Size.Y)}
		case layout.NW:
			offset = f32.Point{Y: float32(-dims2.Size.Y)}
		}
		call := macro.Stop()
		op.Offset(offset).Add(gtx.Ops)
		op.Defer(gtx.Ops, call)
	}
	return dims1
}

func (cf *ColorSelection) event() {
	for range cf.clicker.Clicks() {
		cf.active = !cf.active
	}
}

// Click programmatically
//func (cf *ColorSelection) Click() {
//	cf.clicker.Click()
//}

func (cf *ColorSelection) SetColor(col color.NRGBA) {
	cf.Input.SetColor(col)
}

func (cf *ColorSelection) Color() color.NRGBA {
	return cf.Input.Color()
}

func (cf *ColorSelection) Changed() bool {
	return cf.Input.Changed()
}
