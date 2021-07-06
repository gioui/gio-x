package colorpicker

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image/color"
	"math"
)

func NewToggle(toggle *widget.Clickable, options ...ColorInput) *Toggle {
	return &Toggle{
		toggle:  toggle,
		options: options,
	}
}

type Toggle struct {
	toggle  *widget.Clickable
	index   int
	options []ColorInput
}

func (t *Toggle) Layout(gtx layout.Context) layout.Dimensions {
	t.events()
	w := t.options[t.index].Layout
	macro := op.Record(gtx.Ops)
	dims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, t.toggle, func(gtx layout.Context) layout.Dimensions {
				return toggleIcon.Layout(gtx, unit.Dp(30))
			})
		}),
		layout.Flexed(1, w))
	call := macro.Stop()
	paint.FillShape(gtx.Ops,
		color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		clip.Rect{Max: dims.Size}.Op())
	call.Add(gtx.Ops)
	return dims
}

func (t *Toggle) Color() color.NRGBA {
	return t.options[t.index].Color()
}

func (t *Toggle) SetColor(col color.NRGBA) {
	t.options[t.index].SetColor(col)
}

func (t *Toggle) Changed() bool {
	return t.options[t.index].Changed()
}

func (t *Toggle) events() {
	for range t.toggle.Clicks() {
		col := t.options[t.index].Color()
		t.index = int(math.Mod(float64(t.index+1), float64(len(t.options))))
		t.options[t.index].SetColor(col)
	}
}
