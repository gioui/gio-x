package outlay

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

// Cell lays out the Table cell located at column x and row y.
type Cell func(gtx layout.Context, x, y int) layout.Dimensions

// Table lays out cells by their coordinates.
// All cells within a column have the same width, and
// the same height within a row.
type Table struct {
	// CellSize returns the size for the cell located at column x and row y.
	CellSize     func(m unit.Metric, x, y int) image.Point
	xList, yList layout.List
	x, y         int
}

func (t *Table) Layout(gtx layout.Context, xn, yn int, el Cell) layout.Dimensions {
	defer op.Save(gtx.Ops).Load()
	csMax := gtx.Constraints.Max

	// In order to deliver the same scroll events for both lists,
	// they are collected by a dedicated InputOp and dispatched to
	// both lists by a dedicated queue. This ensures that only the
	// lists receive the pointer events that are collected for scrolling
	// and the widgets displayed within cells are not polluted by them.

	// Collect the scrolling events for the Table. See the dedicated InputOp below.
	scrollEvents := gtx.Events(t)
	// For each list, use a copy of the context with a dedicated queue that only returns
	// the scroll events.
	listContext := gtx
	listContext.Queue = queue(scrollEvents)

	t.xList.Axis = layout.Horizontal
	t.xList.Layout(listContext, xn, func(gtx layout.Context, x int) layout.Dimensions {
		sz := t.CellSize(gtx.Metric, x, t.y)
		return layout.Dimensions{Size: image.Point{X: sz.X, Y: csMax.Y}}
	})
	t.x = t.xList.Position.First

	t.yList.Axis = layout.Vertical
	t.yList.Layout(listContext, yn, func(gtx layout.Context, y int) layout.Dimensions {
		sz := t.CellSize(gtx.Metric, t.x, y)
		return layout.Dimensions{Size: image.Point{X: csMax.X, Y: sz.Y}}
	})
	t.y = t.yList.Position.First

	// Grab all scroll events for the lists.
	pointer.Rect(image.Rectangle{
		Max: gtx.Constraints.Max,
	}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   t,
		Types: pointer.Press | pointer.Drag | pointer.Release | pointer.Scroll,
	}.Add(gtx.Ops)

	// Offset the start position for truncated last columns and rows.
	clip.Rect(image.Rectangle{Max: csMax}).Add(gtx.Ops)
	p := image.Point{
		X: -t.xList.Position.Offset,
		Y: -t.yList.Position.Offset,
	}
	op.Offset(layout.FPt(p)).Add(gtx.Ops)

	gtx.Constraints.Min = image.Point{}
	var yy int
	for y := t.y; y < yn; y++ {
		var xx int
		var sz image.Point
		for x := t.x; x < xn; x++ {
			sz = t.CellSize(gtx.Metric, x, y)
			gtx.Constraints.Max = sz
			// For cells, use the supplied context and its queue.
			el(gtx, x, y)
			if xx >= csMax.X {
				yy += sz.Y
				break
			}
			xx += sz.X
			op.Offset(f32.Point{X: float32(sz.X)}).Add(gtx.Ops)
		}
		if yy >= csMax.Y {
			break
		}
		pt := image.Point{X: -xx, Y: sz.Y}
		op.Offset(layout.FPt(pt)).Add(gtx.Ops)
	}
	return layout.Dimensions{Size: csMax}
}

// queue provides an event queue delivering the same events for any tag.
type queue []event.Event

func (q queue) Events(_ event.Tag) []event.Event { return q }
