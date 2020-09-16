package scroll

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Bar represents a scrolling indicator for a layout.List
type Bar struct {
	color.RGBA
}

// DefaultBar returns a bar with a translucent gray background.
func DefaultBar() Bar {
	return Bar{color.RGBA{A: 200}}
}

// Layout renders the bar on the right edge of the current allowed Constraints.
// The provided position should be the current scroll position of the layout.List
// that the scroll bar is indicating progress for, and the length should be the
// number of elements in that layout.List.
func (b Bar) Layout(gtx layout.Context, pos layout.Position, length int) layout.Dimensions {
	progress := float32(pos.First) / float32(length)
	return layout.NE.Layout(gtx, func(gtx C) D {
		indicatorHeightDp := unit.Dp(16)
		indicatorHeightPx := gtx.Px(indicatorHeightDp)
		heightDp := float32(gtx.Constraints.Max.Y) / gtx.Metric.PxPerDp
		width := gtx.Px(unit.Dp(8))
		top := unit.Dp(heightDp * progress)
		if top.V+indicatorHeightDp.V > heightDp {
			top = unit.Dp(heightDp - indicatorHeightDp.V)
		}
		radii := float32(gtx.Px(unit.Dp(4)))
		return layout.Inset{
			Top:    top,
			Right:  unit.Dp(2),
			Bottom: unit.Dp(2),
		}.Layout(gtx, func(gtx C) D {
			size := f32.Point{X: float32(width), Y: float32(indicatorHeightPx)}
			return DrawRect(gtx, b.RGBA, size, radii)
		})
	})
}

// DrawRect creates a rectangle of the provided background color with
// Dimensions specified by size and a corner radius (on all corners)
// specified by radii.
func DrawRect(gtx C, background color.RGBA, size f32.Point, radii float32) D {
	stack := op.Push(gtx.Ops)
	paintOp := paint.ColorOp{Color: background}
	paintOp.Add(gtx.Ops)
	bounds := f32.Rectangle{
		Max: size,
	}
	clip.RRect{
		Rect: bounds,
		NW:   radii,
		NE:   radii,
		SE:   radii,
		SW:   radii,
	}.Add(gtx.Ops)
	paint.PaintOp{
		Rect: bounds,
	}.Add(gtx.Ops)
	stack.Pop()
	return layout.Dimensions{Size: image.Pt(int(size.X), int(size.Y))}
}
