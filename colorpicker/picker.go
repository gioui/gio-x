package colorpicker

// https://bgrins.github.io/spectrum/#why

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
)

func NewPicker(th *material.Theme) *Picker {
	cp := &Picker{
		tone:  &Position{},
		hue:   &widget.Float{Axis: layout.Horizontal},
		theme: th}
	return cp
}

type Picker struct {
	// Encode hsv saturation on X-axis and hsv value on y-axis.
	tone  *Position
	hue   *widget.Float
	hsv   HSVColor
	alpha byte
	theme *material.Theme
}

func (p *Picker) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(p.layoutGradiants),
		layout.Rigid(p.layoutRainbow))
}

func (p *Picker) layoutGradiants(gtx layout.Context) layout.Dimensions {
	w := gtx.Constraints.Max.X
	h := gtx.Px(unit.Dp(120))
	dr := image.Rectangle{Max: image.Point{X: w, Y: h}}
	primary := HsvToRgb(HSVColor{p.hue.Value * 360, 1, 1})
	stack := op.Save(gtx.Ops)
	topRight := f32.Point{X: float32(dr.Max.X), Y: float32(dr.Min.Y)}
	topLeft := f32.Point{X: float32(dr.Min.X), Y: float32(dr.Min.Y)}
	bottomRight := f32.Point{X: float32(dr.Max.X), Y: float32(dr.Max.Y)}
	paint.LinearGradientOp{
		Stop1:  topRight,
		Stop2:  topLeft,
		Color1: primary,
		Color2: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	}.Add(gtx.Ops)
	clip.Rect(dr).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	paint.LinearGradientOp{
		Stop1:  topRight,
		Stop2:  bottomRight,
		Color1: color.NRGBA{},
		Color2: color.NRGBA{R: 0x00, G: 0x00, B: 0x0, A: 0xff},
	}.Add(gtx.Ops)
	clip.Rect(dr).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Load()

	gtx.Constraints = layout.Exact(image.Point{X: w, Y: h})
	p.tone.Layout(gtx, 1, f32.Point{}, f32.Point{X: 1, Y: 1})
	drawControl(p.tone.Pos(), 10, 1, gtx)

	return layout.Dimensions{Size: dr.Max}
}

func (p *Picker) layoutRainbow(gtx layout.Context) layout.Dimensions {
	w := gtx.Constraints.Max.X
	h := gtx.Px(unit.Dp(20))
	gtx.Constraints = layout.Exact(image.Point{X: w, Y: h})
	drawRainbow(gtx)
	p.hue.Layout(gtx, 1, 0, 1)
	drawControl(f32.Point{p.hue.Pos(), float32(h / 2)}, 10, 1, gtx)
	return layout.Dimensions{Size: image.Point{X: w, Y: h}}
}

func (p *Picker) SetColor(col color.NRGBA) {
	p.hsv = RgbToHsv(col)
	p.tone.X = p.hsv.S
	p.tone.Y = 1 - p.hsv.V
	p.hue.Value = p.hsv.H / 360
	p.alpha = col.A
}

func (p *Picker) Color() color.NRGBA {
	col := HsvToRgb(p.hsv)
	col.A = p.alpha
	return col
}

func (p *Picker) Changed() bool {
	changed := false
	if p.tone.Changed() {
		changed = true
		p.hsv.S = p.tone.X
		p.hsv.V = 1 - p.tone.Y
	} else if p.hue.Changed() {
		changed = true
		p.hsv.H = p.hue.Value * 360
	}
	return changed
}
