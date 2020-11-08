package scroll

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Scrollable holds state of a scrolling widget. The Scrolled() method is
// used to tell both whether a scroll operation occurred during the last frame
// as well as the progress through the scrollable region at the end of the
// scroll operation.
type Scrollable struct {
	// Track clicks.
	clickable widget.Clickable
	// Track drag events.
	drag gesture.Drag
	// Has the bar scrolled since the previous frame?
	scrolled bool
	// Cached length of scroll region after layout has been computed. This can be
	// off if the screen is being resized, but we have no better way to acquire
	// this data.
	length int
	// progress is how far along we are as a fraction between 0 and 1.
	progress float32
}

// Bar represents a scrolling indicator for a layout.List
type Bar struct {
	*Scrollable
	// Color of the scroll indicator.
	Color color.RGBA
	// Progress tells the bar where to render the indicator as a fraction [0, 1].
	Progress float32
	// Scale tells the bar what fraction of the available axis space it should
	// occupy as a fraction between [0, 1].
	Scale float32
	// Axis along which the bar is oriented.
	Axis Axis
	// Axis independent size.
	Thickness unit.Value
	// MinLength is the minimum length of the scroll indicator. Regardless of
	// the scale of the bar, it will not be displayed shorter than this. If
	// the scale parameter isn't provided, the indicator will always have
	// this length.
	MinLength unit.Value
}

// Axis specifies the scroll bar orientation.
// Default to `Vertical`.
type Axis int

const (
	Vertical   = 0
	Horizontal = 1
)

// DefaultBar returns a bar with a translucent gray background. The progress
// parameter tells the bar how far through its range of motion to draw itself.
// The scale parameter tells the bar what fraction of the scrollable space is
// visible. Scale may be left as zero to use a minimum-length scroll indicator
// that does not respond to changes in the length of the scrollable region.
func DefaultBar(state *Scrollable, progress, scale float32) Bar {
	return Bar{
		Scrollable: state,
		Progress:   progress,
		Scale:      scale,
		Color:      color.RGBA{A: 200},
		Thickness:  unit.Dp(8),
		MinLength:  unit.Dp(16),
	}
}

// Update the internal state of the bar.
func (sb *Scrollable) Update(gtx C, axis Axis) {
	sb.scrolled = false
	// Restrict progress to [0, 1].
	defer func() {
		if sb.progress > 1 {
			sb.progress = 1
		} else if sb.progress < 0 {
			sb.progress = 0
		}
	}()
	pickAxis := func(pt f32.Point) (v float32) {
		switch axis {
		case Vertical:
			v = pt.Y
		case Horizontal:
			v = pt.X
		}
		return v
	}
	if sb.clickable.Clicked() {
		if presses := sb.clickable.History(); len(presses) > 0 {
			press := presses[len(presses)-1]
			sb.progress = float32(pickAxis(press.Position)) / float32(sb.length)
			sb.scrolled = true
		}
	}
	if drags := sb.drag.Events(gtx.Metric, gtx, axis.ToGesture()); len(drags) > 0 {
		delta := pickAxis(drags[len(drags)-1].Position)
		sb.progress = (sb.progress*float32(sb.length) + (delta / 2)) / float32(sb.length)
		sb.scrolled = true
	}
}

// Scrolled returns true if the scroll position changed within the last frame.
func (sb Scrollable) Scrolled() (didScroll bool, progress float32) {
	return sb.scrolled, sb.progress
}

// Layout renders the bar into the provided context.
func (sb Bar) Layout(gtx C) D {
	sb.Scrollable.progress = sb.Progress
	sb.Update(gtx, sb.Axis)
	if scrolled, _ := sb.Scrolled(); scrolled {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	scaledLength := float32(0)
	switch sb.Axis {
	case Horizontal:
		scaledLength = (sb.Scale * float32(gtx.Constraints.Max.X))
	case Vertical:
		scaledLength = (sb.Scale * float32(gtx.Constraints.Max.Y))
	}
	if int(scaledLength) > gtx.Px(sb.MinLength) {
		sb.MinLength = unit.Dp(scaledLength / gtx.Metric.PxPerDp)
	}
	return sb.Axis.Layout(gtx, func(gtx C) D {
		if sb.MinLength == (unit.Value{}) {
			sb.MinLength = unit.Dp(16)
		}
		if sb.Thickness == (unit.Value{}) {
			sb.Thickness = unit.Dp(8)
		}
		var (
			total float32
			size  f32.Point
			top   = unit.Dp(2)
			left  = unit.Dp(2)
		)
		switch sb.Axis {
		case Horizontal:
			sb.length = gtx.Constraints.Max.X
			size = f32.Point{
				X: float32(gtx.Px(sb.MinLength)),
				Y: float32(gtx.Px(sb.Thickness)),
			}
			total = float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp
			left = unit.Dp(total * sb.Progress)
			if left.V+sb.MinLength.V > total {
				left = unit.Dp(total - sb.MinLength.V)
			}
		case Vertical:
			sb.length = gtx.Constraints.Max.Y
			size = f32.Point{
				X: float32(gtx.Px(sb.Thickness)),
				Y: float32(gtx.Px(sb.MinLength)),
			}
			total = float32(gtx.Constraints.Max.Y) / gtx.Metric.PxPerDp
			top = unit.Dp(total * sb.Progress)
			if top.V+sb.MinLength.V > total {
				top = unit.Dp(total - sb.MinLength.V)
			}
		}
		return clickBox(gtx, &sb.clickable, func(gtx C) D {
			barAreaDims := layout.Inset{
				Top:    top,
				Right:  unit.Dp(2),
				Left:   left,
				Bottom: unit.Dp(2),
			}.Layout(gtx, func(gtx C) D {
				pointer.Rect(image.Rectangle{
					Max: image.Point{
						X: int(size.X),
						Y: int(size.Y),
					},
				}).Add(gtx.Ops)
				sb.drag.Add(gtx.Ops)
				return rect{
					Color: sb.Color,
					Size:  size,
					Radii: float32(gtx.Px(unit.Dp(4))),
				}.Layout(gtx)
			})
			switch sb.Axis {
			case Vertical:
				barAreaDims.Size.Y = gtx.Constraints.Max.Y
			case Horizontal:
				barAreaDims.Size.X = gtx.Constraints.Max.X
			}
			return barAreaDims
		})
	})
}

func (axis Axis) Layout(gtx C, widget layout.Widget) D {
	if axis == Vertical {
		return layout.NE.Layout(gtx, widget)
	}
	if axis == Horizontal {
		return layout.SW.Layout(gtx, widget)
	}
	return layout.Dimensions{}
}

func (axis Axis) ToGesture() (g gesture.Axis) {
	switch axis {
	case Vertical:
		g = gesture.Vertical
	case Horizontal:
		g = gesture.Horizontal
	}
	return g
}

// rect creates a rectangle of the provided background color with
// Dimensions specified by size and a corner radius (on all corners)
// specified by radii.
type rect struct {
	Color color.RGBA
	Size  f32.Point
	Radii float32
}

// Layout renders the Rect into the provided context
func (r rect) Layout(gtx C) D {
	return drawRect(gtx, r.Color, r.Size, r.Radii)
}

// drawRect creates a rectangle of the provided background color with
// Dimensions specified by size and a corner radius (on all corners)
// specified by radii.
func drawRect(gtx C, background color.RGBA, size f32.Point, radii float32) D {
	bounds := f32.Rectangle{
		Max: size,
	}
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(bounds, radii).Op(gtx.Ops))
	return layout.Dimensions{Size: image.Pt(int(size.X), int(size.Y))}
}

// clickBox lays out a rectangular clickable widget without further
// decoration.
func clickBox(gtx layout.Context, button *widget.Clickable, w layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(button.Layout),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			clip.RRect{
				Rect: f32.Rectangle{Max: f32.Point{
					X: float32(gtx.Constraints.Min.X),
					Y: float32(gtx.Constraints.Min.Y),
				}},
			}.Add(gtx.Ops)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}
