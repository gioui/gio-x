package component

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

// Resize provides a draggable handle in between two widgets for resizing their area.
type Resize struct {
	// Axis defines how the widgets and the handle are laid out.
	Axis layout.Axis
	// Ratio defines how much space is available to the first widget.
	Ratio float32
	float float
}

// Layout displays w1 and w2 with handle in between.
//
// The widgets w1 and w2 must be able to gracefully resize their minimum and maximum dimensions
// in order for the resize to be smooth.
func (rs *Resize) Layout(gtx layout.Context, w1, w2, handle layout.Widget) layout.Dimensions {
	// Compute the first widget's max width/height.
	rs.float.Length = rs.Axis.Convert(gtx.Constraints.Max).X
	rs.float.Pos = int(rs.Ratio * float32(rs.float.Length))
	m := op.Record(gtx.Ops)
	dims := rs.float.Layout(gtx, rs.Axis, handle)
	c := m.Stop()
	rs.Ratio = float32(rs.float.Pos) / float32(rs.float.Length)
	return layout.Flex{
		Axis: rs.Axis,
	}.Layout(gtx,
		layout.Flexed(rs.Ratio, w1),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			c.Add(gtx.Ops)
			return dims
		}),
		layout.Flexed(1-rs.Ratio, w2),
	)
}

type float struct {
	Length int // max constraint for the axis
	Pos    int // position in pixels of the handle
	drag   gesture.Drag
}

func (f *float) Layout(gtx layout.Context, axis layout.Axis, w layout.Widget) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	dims := w(gtx)

	var de *pointer.Event
	for _, e := range f.drag.Events(gtx.Metric, gtx, gesture.Axis(axis)) {
		if e.Type == pointer.Drag {
			de = &e
		}
	}
	if de != nil {
		xy := de.Position.X
		if axis == layout.Vertical {
			xy = de.Position.Y
		}
		f.Pos += int(xy)
	}

	// Clamp the handle position, leaving it always visible.
	if f.Pos < 0 {
		f.Pos = 0
	} else if f.Pos > f.Length {
		f.Pos = f.Length
	}

	defer op.Save(gtx.Ops).Load()
	rect := image.Rectangle{Max: dims.Size}
	pointer.Rect(rect).Add(gtx.Ops)
	f.drag.Add(gtx.Ops)
	cursor := pointer.CursorRowResize
	if axis == layout.Horizontal {
		cursor = pointer.CursorColResize
	}
	pointer.CursorNameOp{Name: cursor}.Add(gtx.Ops)

	return layout.Dimensions{Size: dims.Size}
}
