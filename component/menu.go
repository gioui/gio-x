package component

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// CardStyle defines the visual aspects of a material design surface
// with (optionally) rounded corners and a drop shadow.
type CardStyle struct {
	*material.Theme
	layout.Inset
	Radius unit.Value
	Shadow color.NRGBA
}

func Card(th *material.Theme) CardStyle {
	defaultShadowSize := unit.Dp(20)
	return CardStyle{
		Theme:  th,
		Inset:  layout.UniformInset(defaultShadowSize),
		Radius: defaultShadowSize,
		Shadow: color.NRGBA{A: 155},
	}
}

func (c CardStyle) Layout(gtx C, w layout.Widget) D {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	finalDims := c.Inset.Layout(gtx, func(gtx C) D {
		call.Add(gtx.Ops)
		return dims
	})
	defer op.Save(gtx.Ops).Load()
	clip.UniformRRect(f32.Rectangle{Max: layout.FPt(finalDims.Size)}, float32(gtx.Px(c.Radius))).Add(gtx.Ops)

	left := (gtx.Px(c.Inset.Left))
	top := (gtx.Px(c.Inset.Top))
	right := (gtx.Px(c.Inset.Right))
	bottom := (gtx.Px(c.Inset.Bottom))

	pathFrom := func(gtx C, points ...f32.Point) clip.PathSpec {
		p := clip.Path{}
		p.Begin(gtx.Ops)
		p.MoveTo(points[0])
		for _, point := range points[1:] {
			p.LineTo(point)
		}
		p.Close()
		return p.End()
	}
	topSpec := pathFrom(gtx,
		f32.Point{},
		f32.Point{
			X: float32(left + dims.Size.X + right),
		},
		f32.Point{
			X: float32(left+dims.Size.X) + .5,
			Y: float32(top),
		},
		f32.Point{
			X: float32(left) + .25,
			Y: float32(top),
		},
	)
	leftSpec := pathFrom(gtx,
		f32.Point{},
		f32.Point{X: float32(left), Y: float32(top) - .5},
		f32.Point{X: float32(left), Y: float32(top+dims.Size.Y) - .25},
		f32.Point{Y: float32(top + dims.Size.Y + bottom)},
	)
	rightSpec := pathFrom(gtx,
		f32.Point{X: float32(left + right + dims.Size.X)},
		f32.Point{X: float32(left + right + dims.Size.X),
			Y: float32(top + bottom + dims.Size.Y)},
		f32.Point{X: float32(left + dims.Size.X),
			Y: float32(top+dims.Size.Y) + .5},
		f32.Point{X: float32(left + dims.Size.X),
			Y: float32(top) + .25},
	)
	bottomSpec := pathFrom(gtx,
		f32.Point{Y: float32(top + bottom + dims.Size.Y)},
		f32.Point{X: float32(left + right + dims.Size.X),
			Y: float32(top + bottom + dims.Size.Y)},
		f32.Point{X: float32(left+dims.Size.X) - .25,
			Y: float32(top + dims.Size.Y)},
		f32.Point{X: float32(left) - .5,
			Y: float32(top + dims.Size.Y)},
	)

	topGrad := paint.LinearGradientOp{
		Stop2:  f32.Point{Y: float32(top)},
		Color2: c.Shadow,
	}
	leftGrad := paint.LinearGradientOp{
		Stop2:  f32.Point{X: float32(left)},
		Color2: c.Shadow,
	}
	rightGrad := paint.LinearGradientOp{
		Stop1:  f32.Point{X: float32(left + right + dims.Size.X)},
		Stop2:  f32.Point{X: float32(left + dims.Size.X)},
		Color2: c.Shadow,
	}
	bottomGrad := paint.LinearGradientOp{
		Stop1:  f32.Point{Y: float32(top + bottom + dims.Size.Y)},
		Stop2:  f32.Point{Y: float32(top + dims.Size.Y)},
		Color2: c.Shadow,
	}
	apply := func(gtx C, grad paint.LinearGradientOp, p clip.PathSpec) {
		defer op.Save(gtx.Ops).Load()
		clip.Outline{Path: p}.Op().Add(gtx.Ops)
		grad.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
	apply(gtx, topGrad, topSpec)
	apply(gtx, leftGrad, leftSpec)
	apply(gtx, rightGrad, rightSpec)
	apply(gtx, bottomGrad, bottomSpec)
	return finalDims
}

// MenuState holds the state of a menu material design component
// across frames.
type MenuState struct {
	OptionList layout.List
	Options    []func(gtx C) D
}

// MenuStyle defines the presentation of a material design menu component.
type MenuStyle struct {
	*MenuState
	*material.Theme
	layout.Inset
	CardStyle
}

func Menu(th *material.Theme, state *MenuState) MenuStyle {
	return MenuStyle{
		Theme:     th,
		MenuState: state,
		CardStyle: Card(th),
	}
}

func (m MenuStyle) Layout(gtx C) D {
	m.OptionList.Axis = layout.Vertical
	return m.CardStyle.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				return Rect{
					Color: m.Theme.Bg,
					Size:  gtx.Constraints.Min,
				}.Layout(gtx)
			}),
			layout.Stacked(func(gtx C) D {
				return m.OptionList.Layout(gtx, len(m.Options), func(gtx C, index int) D {
					return m.Options[index](gtx)
				})
			}),
		)
	})
}
