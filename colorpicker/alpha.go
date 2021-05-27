package colorpicker

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
	"image/color"
)

func NewAlphaSlider() *AlphaSlider {
	return &AlphaSlider{}
}

type AlphaSlider struct {
	slider widget.Float
	color  color.NRGBA
}

func (s *AlphaSlider) Color() color.NRGBA {
	return s.color
}

func (s *AlphaSlider) SetColor(col color.NRGBA) {
	s.color = col
	s.slider.Value = float32(s.color.A / 255)
}

func (s *AlphaSlider) Layout(gtx layout.Context) layout.Dimensions {
	w := gtx.Constraints.Max.X
	h := gtx.Px(unit.Dp(20))

	gtx.Constraints = layout.Exact(image.Point{w, h})
	drawCheckerboard(gtx)

	col1 := s.Color()
	col2 := col1
	col1.A = 0x00
	col2.A = 0xff
	defer op.Save(gtx.Ops).Load()
	paint.LinearGradientOp{
		Stop1:  f32.Point{float32(0), 0},
		Stop2:  f32.Point{float32(w), 0},
		Color1: col1,
		Color2: col2,
	}.Add(gtx.Ops)
	dr := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: w, Y: h}}
	clip.Rect(dr).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	s.slider.Layout(gtx, 1, 0, 1)
	x := s.slider.Pos()
	drawControl(f32.Point{x, float32(h / 2)}, 10, 1, gtx)

	return layout.Dimensions{Size: image.Point{X: w, Y: h}}
}

func (s *AlphaSlider) Changed() bool {
	if !s.slider.Changed() {
		return false
	}
	s.color.A = byte(s.slider.Value * 255)
	return true
}
