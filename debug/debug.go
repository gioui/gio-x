// Package debug provides tools for layout debugging.
package debug

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// dragBox holds state for a draggable, hoverable rectangle with drag that can
// accumulate.
type dragBox struct {
	committedDrag    image.Point
	activeDrag       image.Point
	activeDragOrigin image.Point
	drag             gesture.Drag
	hover            gesture.Hover
}

// Add inserts the dragBox's input operations into the ops list.
func (d *dragBox) Add(ops *op.Ops) {
	d.drag.Add(ops)
	d.hover.Add(ops)
}

// Update processes events from the queue using the given metric and updates the
// drag position.
func (d *dragBox) Update(metric unit.Metric, queue event.Queue) {
	for _, ev := range d.drag.Events(metric, queue, gesture.Both) {
		switch ev.Type {
		case pointer.Press:
			d.activeDragOrigin = ev.Position.Round()
		case pointer.Cancel, pointer.Release:
			d.activeDragOrigin = image.Point{}
			d.committedDrag = d.committedDrag.Add(d.activeDrag)
			d.activeDrag = image.Point{}
		case pointer.Drag:
			d.activeDrag = d.activeDragOrigin.Sub(ev.Position.Round())
		}
	}
}

// CurrentDrag returns the current accumulated drag (both drag from previous events
// and any in-progress drag gesture).
func (d *dragBox) CurrentDrag() image.Point {
	return d.committedDrag.Add(d.activeDrag)
}

// Active returns whether the user is hovering or interacting with the dragBox.
func (d *dragBox) Active(queue event.Queue) bool {
	return d.drag.Dragging() || d.drag.Pressed() || d.hover.Hovered(queue)
}

// Reset clears all accumulated drag.
func (d *dragBox) Reset() {
	d.committedDrag = image.Point{}
}

// ConstraintEditor provides controls to edit layout constraints live.
type ConstraintEditor struct {
	maxBox dragBox
	minBox dragBox
}

func outline(ops *op.Ops, width int, area image.Point) clip.PathSpec {
	widthF := float32(width)
	var builder clip.Path
	builder.Begin(ops)
	builder.LineTo(f32.Pt(float32(area.X), 0))
	builder.LineTo(f32.Pt(float32(area.X), float32(area.Y)))
	builder.LineTo(f32.Pt(0, float32(area.Y)))
	builder.LineTo(f32.Pt(0, 0))
	if area.X > width && area.Y > width {
		builder.MoveTo(f32.Pt(widthF, widthF))
		builder.LineTo(f32.Pt(widthF, float32(area.Y)-widthF))
		builder.LineTo(f32.Pt(float32(area.X)-widthF, float32(area.Y)-widthF))
		builder.LineTo(f32.Pt(float32(area.X)-widthF, widthF))
		builder.LineTo(f32.Pt(widthF, widthF))
	}
	builder.Close()
	return builder.End()
}

// Wrap returns a new layout function with the ConstraintEditor wrapping the provided widget.
// This can make inserting the editor more ergonomic than the Layout signature.
func (c *ConstraintEditor) Wrap(w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return c.Layout(gtx, w)
	}
}

// Layout the constraint editor to debug the layout of w.
func (c *ConstraintEditor) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	originalConstraints := gtx.Constraints
	gtx.Constraints = gtx.Constraints.SubMax(c.maxBox.CurrentDrag())
	gtx.Constraints = gtx.Constraints.AddMin(image.Point{}.Sub(c.minBox.CurrentDrag()))
	if gtx.Constraints.Max.X < gtx.Dp(5) || gtx.Constraints.Max.Y < gtx.Dp(5) {
		gtx.Constraints = originalConstraints
		c.maxBox.Reset()
		c.minBox.Reset()
	}
	dims := w(gtx)
	oneDp := gtx.Dp(1)
	minSpec := outline(gtx.Ops, oneDp, gtx.Constraints.Min)
	maxSpec := outline(gtx.Ops, oneDp, gtx.Constraints.Max)
	sizeSpec := outline(gtx.Ops, oneDp, dims.Size)
	rec := op.Record(gtx.Ops)
	// Display the static widget size.
	paint.FillShape(gtx.Ops, color.NRGBA{G: 255, A: 150}, clip.Outline{Path: sizeSpec}.Op())
	paint.FillShape(gtx.Ops, color.NRGBA{G: 255, A: 50}, clip.Rect{Max: dims.Size}.Op())
	// Display the interactive max constraint controls.
	paint.FillShape(gtx.Ops, color.NRGBA{R: 255, A: 150}, clip.Outline{Path: maxSpec}.Op())
	if c.maxBox.Active(gtx.Queue) {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 255, A: 50}, clip.Rect{Max: gtx.Constraints.Max}.Op())
	}

	maxDragArea := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	c.maxBox.Add(gtx.Ops)
	maxDragArea.Pop()

	// Display the interactive min constraint controls.
	paint.FillShape(gtx.Ops, color.NRGBA{B: 255, A: 150}, clip.Outline{Path: minSpec}.Op())
	if c.minBox.Active(gtx.Queue) {
		paint.FillShape(gtx.Ops, color.NRGBA{B: 255, A: 50}, clip.Rect{Max: gtx.Constraints.Min}.Op())
	}

	minDragArea := clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops)
	c.minBox.Add(gtx.Ops)
	minDragArea.Pop()

	op.Defer(gtx.Ops, rec.Stop())

	c.maxBox.Update(gtx.Metric, gtx.Queue)
	c.minBox.Update(gtx.Metric, gtx.Queue)
	return dims
}
