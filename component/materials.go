package component

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type Rect struct {
	Color color.NRGBA
	Size  image.Point
	Radii float32
}

func (r Rect) Layout(gtx C) D {
	paint.FillShape(
		gtx.Ops,
		r.Color,
		clip.UniformRRect(
			f32.Rectangle{
				Max: layout.FPt(r.Size),
			},
			r.Radii,
		).Op(gtx.Ops))
	return layout.Dimensions{Size: image.Pt(r.Size.X, r.Size.Y)}
}
