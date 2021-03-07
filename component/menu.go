package component

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type ShadowStyle struct {
	Shape     clip.RRect
	Elevation unit.Value
	Darkest   color.NRGBA
}

func Shadow(rect clip.RRect, elevation unit.Value) ShadowStyle {
	return ShadowStyle{
		Elevation: elevation,
		Shape:     rect,
		Darkest:   color.NRGBA{A: 200},
	}
}

func (s ShadowStyle) radius(shapeRR float32) float32 {
	return shapeRR
}

func (s ShadowStyle) spread(gtx C) float32 {
	return float32(gtx.Px(s.Elevation))
}

func (s ShadowStyle) Layout(gtx C) D {
	nwRR := s.radius(s.Shape.NW)
	neRR := s.radius(s.Shape.NE)
	seRR := s.radius(s.Shape.SE)
	swRR := s.radius(s.Shape.SW)
	spread := s.spread(gtx)

	shadowW := s.Shape.Rect.Dx() + 2*spread
	shadowH := s.Shape.Rect.Dy() + 2*spread

	// topW := shadowW - nwRR - neRR
	// leftH := shadowH - nwRR - swRR
	// rightH := shadowH - neRR - seRR
	// bottomW := shadowW - seRR - swRR

	// NW corner.
	saved := op.Save(gtx.Ops)
	paint.RadialGradientOp{
		Stop1: f32.Point{
			X: nwRR,
			Y: nwRR,
		},
		Color1: s.Darkest,
		Stop2: f32.Point{
			X: nwRR - spread,
			Y: nwRR,
		},
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	saved.Load()
	// NE corner.
	paint.RadialGradientOp{
		Stop1: f32.Point{
			X: shadowW - neRR,
			Y: neRR,
		},
		Color1: s.Darkest,
		Stop2: f32.Point{
			X: shadowW - neRR + spread,
			Y: neRR,
		},
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	// SE corner.
	paint.RadialGradientOp{
		Stop1: f32.Point{
			X: shadowW - seRR,
			Y: shadowH - seRR,
		},
		Color1: s.Darkest,
		Stop2: f32.Point{
			X: shadowW - seRR + spread,
			Y: shadowH - seRR,
		},
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	// SW corner.
	paint.RadialGradientOp{
		Stop1: f32.Point{
			X: swRR,
			Y: shadowH - swRR,
		},
		Color1: s.Darkest,
		Stop2: f32.Point{
			X: swRR - spread,
			Y: shadowH - swRR,
		},
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return D{Size: image.Pt(int(shadowW), int(shadowH))}
}

// CardStyle defines the visual aspects of a material design surface
// with (optionally) rounded corners and a drop shadow.
type CardStyle struct {
	*material.Theme
	Radius    unit.Value
	Elevation unit.Value
	Shadow    color.NRGBA
}

func Card(th *material.Theme) CardStyle {
	return CardStyle{
		Theme:     th,
		Radius:    unit.Dp(0),
		Elevation: unit.Dp(8),
		Shadow:    color.NRGBA{A: 155},
	}
}

func (c CardStyle) Layout(gtx C, w layout.Widget) D {
	inset := layout.UniformInset(c.Elevation.Scale(.5))
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()
	dimsCopy := dims

	elevationAdjustment := gtx.Px(c.Elevation)

	// Adjust dims to account for the part of the drop shadow that will
	// be underneath the card.
	dims.Size.X -= 2 * elevationAdjustment
	dims.Size.Y -= 2 * elevationAdjustment

	left := (gtx.Px(inset.Left) + elevationAdjustment)
	top := (gtx.Px(inset.Top) + elevationAdjustment)
	right := (gtx.Px(inset.Right) + elevationAdjustment)
	bottom := (gtx.Px(inset.Bottom) + elevationAdjustment)

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
	func() {
		defer op.Save(gtx.Ops).Load()
		op.Offset(f32.Pt(0, float32(gtx.Px(unit.Dp(2))/2))).Add(gtx.Ops)
		apply(gtx, topGrad, topSpec)
		apply(gtx, leftGrad, leftSpec)
		apply(gtx, rightGrad, rightSpec)
		apply(gtx, bottomGrad, bottomSpec)
	}()
	finalDims := inset.Layout(gtx, func(gtx C) D {
		defer op.Save(gtx.Ops).Load()
		clip.UniformRRect(f32.Rectangle{Max: layout.FPt(dimsCopy.Size)}, float32(gtx.Px(c.Radius))).Add(gtx.Ops)
		call.Add(gtx.Ops)
		return dimsCopy
	})
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
