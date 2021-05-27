package colorpicker

import (
	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"image"
)

type Position struct {
	X float32
	Y float32

	drag          gesture.Drag
	x             float32 // position normalized to [0, 1]
	y             float32 // position normalized to [0, 1]
	width, height float32
	changed       bool
}

func (p *Position) Layout(gtx layout.Context, pointerMargin int, min, max f32.Point) layout.Dimensions {
	size := gtx.Constraints.Min
	p.width = float32(size.X)
	p.height = float32(size.Y)
	var de *pointer.Event
	for _, e := range p.drag.Events(gtx.Metric, gtx.Queue, gesture.Horizontal) {
		if e.Type == pointer.Press || e.Type == pointer.Drag {
			de = &e
		}
	}

	x, y := p.X, p.Y
	if de != nil {
		p.x = de.Position.X / p.width
		p.y = de.Position.Y / p.height
		x = min.X + (max.X-min.X)*p.x
		y = min.Y + (max.Y-min.Y)*p.y
	} else if min != max {
		p.x = x/(max.X-min.X) - min.X
		p.y = y/(max.Y-min.Y) - min.Y
	}
	// Unconditionally call setValue in case min, max, or value changed.
	p.setValue(x, y, min, max)

	if p.x < 0 {
		p.x = 0
	} else if p.x > 1 {
		p.x = 1
	}
	if p.y < 0 {
		p.y = 0
	} else if p.y > 1 {
		p.y = 1
	}

	defer op.Save(gtx.Ops).Load()
	//margin := image.Pt(pointerMargin, pointerMargin)
	//rect := image.Rectangle{
	//	Min: margin.Mul(-1),
	//	Max: size.Add(margin),
	//}
	pointer.Rect(image.Rectangle{Max: size}).Add(gtx.Ops)
	p.drag.Add(gtx.Ops)

	return layout.Dimensions{Size: size}
}

func (p *Position) setValue(x, y float32, min, max f32.Point) {
	//if min > max {
	//	min, max = max, min
	//} ??
	if x < min.X {
		x = min.X
	} else if x > max.X {
		x = max.X
	}
	if y < min.Y {
		y = min.Y
	} else if y > max.Y {
		y = max.Y
	}
	if p.X != x {
		p.X = x
		p.changed = true
	}
	if p.Y != y {
		p.Y = y
		p.changed = true
	}
}

// Pos reports the selected position.
func (p *Position) Pos() f32.Point {
	return f32.Point{X: p.x * p.width, Y: p.y * p.height}
}

// Changed reports whether the value has changed since
// the last call to Changed.
func (p *Position) Changed() bool {
	changed := p.changed
	p.changed = false
	return changed
}
