package colorpicker

import (
	"encoding/hex"
	"image"
	"image/color"
	"log"
	"strconv"
	"strings"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type State struct {
	R, G, B, A widget.Float
	widget.Editor

	changed bool
}

func (s *State) SetColor(c color.RGBA) {
	s.R.Value = float32(c.R) / 255.0
	s.G.Value = float32(c.G) / 255.0
	s.B.Value = float32(c.B) / 255.0
	s.A.Value = float32(c.A) / 255.0
	s.updateEditor()
}

func (s State) Color() color.RGBA {
	return color.RGBA{
		R: s.Red(),
		G: s.Green(),
		B: s.Blue(),
		A: s.Alpha(),
	}
}

func (s State) Red() uint8 {
	return uint8(s.R.Value * 255)
}

func (s State) Green() uint8 {
	return uint8(s.G.Value * 255)
}

func (s State) Blue() uint8 {
	return uint8(s.B.Value * 255)
}

func (s State) Alpha() uint8 {
	return uint8(s.A.Value * 255)
}

func (s State) Changed() bool {
	return s.changed
}

func (s *State) Layout(gtx layout.Context) layout.Dimensions {
	s.changed = false
	if s.R.Changed() || s.G.Changed() || s.B.Changed() || s.A.Changed() {
		s.updateEditor()
	}
	if events := s.Editor.Events(); len(events) != 0 {
		out, err := hex.DecodeString(s.Editor.Text())
		if err == nil && len(out) == 3 {
			s.R.Value = (float32(out[0]) / 255.0)
			s.G.Value = (float32(out[1]) / 255.0)
			s.B.Value = (float32(out[2]) / 255.0)
			s.changed = true
		}
	}
	return layout.Dimensions{}
}

func (s *State) updateEditor() {
	s.Editor.SetText(hex.EncodeToString([]byte{s.Red(), s.Green(), s.Blue()}))
	s.changed = true
}

type PickerStyle struct {
	*State
	*material.Theme
	Label string
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func Picker(th *material.Theme, state *State, label string) PickerStyle {
	return PickerStyle{
		Theme: th,
		State: state,
		Label: label,
	}
}

func (p PickerStyle) Layout(gtx layout.Context) layout.Dimensions {
	p.State.Layout(gtx)
	stack := op.Push(gtx.Ops)
	width := gtx.Constraints.Max.X / 2
	gtx.Constraints = layout.Exact(image.Pt(width, gtx.Constraints.Max.Y))
	sliderMacro := op.Record(gtx.Ops)
	sliderDims := p.layoutSliders(gtx)
	log.Printf("dims: %v", sliderDims)
	slider := sliderMacro.Stop()
	stack.Pop()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			gtx.Constraints = layout.Exact(sliderDims.Size)

			return p.layoutLeftPane(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			slider.Add(gtx.Ops)
			return sliderDims
		}),
	)
}

func (p PickerStyle) layoutLeftPane(gtx C) D {
	inset := layout.UniformInset(unit.Dp(8))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				return material.Body1(p.Theme, p.Label).Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				return material.Editor(p.Theme, &p.Editor, "rrggbb").Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				stack := op.Push(gtx.Ops)
				paint.ColorOp{Color: p.State.Color()}.Add(gtx.Ops)
				paint.PaintOp{
					Rect: f32.Rectangle{
						Max: f32.Point{
							X: float32(gtx.Constraints.Max.X),
							Y: float32(gtx.Constraints.Max.Y),
						},
					},
				}.Add(gtx.Ops)
				stack.Pop()
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
		}),
	)
}

func (p PickerStyle) layoutSliders(gtx C) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.R, "R:", valueString(p.Red()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.G, "G:", valueString(p.Green()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.B, "B:", valueString(p.Blue()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.A, "A:", valueString(p.Alpha()))
		}),
	)
}

func valueString(in uint8) string {
	s := strconv.Itoa(int(in))
	delta := 3 - len(s)
	if delta > 0 {
		s = strings.Repeat(" ", delta) + s
	}
	return s
}

func (p PickerStyle) layoutSlider(gtx C, value *widget.Float, label, valueStr string) D {
	inset := layout.UniformInset(unit.Dp(8))
	layoutDims := layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Body1(p.Theme, label)
				label.Font.Variant = "Mono"
				return label.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			sliderDims := inset.Layout(gtx, material.Slider(p.Theme, value, 0, 1).Layout)
			return sliderDims
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Body1(p.Theme, valueStr)
				label.Font.Variant = "Mono"
				return label.Layout(gtx)
			})
		}),
	)
	return layoutDims
}
