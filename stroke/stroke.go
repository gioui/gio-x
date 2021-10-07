// SPDX-License-Identifier: Unlicense OR MIT

// Package stroke converts complex strokes to gioui.org/op/clip operations.
package stroke

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Path defines the shape of a Stroke.
type Path struct {
	Segments []Segment
}

type Segment struct {
	// op is the operator.
	op segmentOp
	// args is up to three (x, y) coordinates.
	args [3]f32.Point
}

// Dashes defines the dash pattern of a Stroke.
type Dashes struct {
	Phase  float32
	Dashes []float32
}

// Stroke defines a stroke.
type Stroke struct {
	Path  Path
	Width float32 // Width of the stroked path.

	// Miter is the limit to apply to a miter joint.
	// The zero Miter disables the miter joint; setting Miter to +∞
	// unconditionally enables the miter joint.
	Miter float32
	Cap   StrokeCap  // Cap describes the head or tail of a stroked path.
	Join  StrokeJoin // Join describes how stroked paths are collated.

	Dashes Dashes
}

type segmentOp uint8

const (
	segOpMoveTo segmentOp = iota
	segOpLineTo
	segOpQuadTo
	segOpCubeTo
)

// StrokeCap describes the head or tail of a stroked path.
type StrokeCap uint8

const (
	// RoundCap caps stroked paths with a round cap, joining the right-hand and
	// left-hand sides of a stroked path with a half disc of diameter the
	// stroked path's width.
	RoundCap StrokeCap = iota

	// FlatCap caps stroked paths with a flat cap, joining the right-hand
	// and left-hand sides of a stroked path with a straight line.
	FlatCap

	// SquareCap caps stroked paths with a square cap, joining the right-hand
	// and left-hand sides of a stroked path with a half square of length
	// the stroked path's width.
	SquareCap
)

// StrokeJoin describes how stroked paths are collated.
type StrokeJoin uint8

const (
	// RoundJoin joins path segments with a round segment.
	RoundJoin StrokeJoin = iota

	// BevelJoin joins path segments with sharp bevels.
	BevelJoin
)

func MoveTo(p f32.Point) Segment {
	s := Segment{
		op: segOpMoveTo,
	}
	s.args[0] = p
	return s
}

func LineTo(p f32.Point) Segment {
	s := Segment{
		op: segOpLineTo,
	}
	s.args[0] = p
	return s
}

func QuadTo(ctrl, end f32.Point) Segment {
	s := Segment{
		op: segOpQuadTo,
	}
	s.args[0] = ctrl
	s.args[1] = end
	return s
}

func CubeTo(ctrl0, ctrl1, end f32.Point) Segment {
	s := Segment{
		op: segOpCubeTo,
	}
	s.args[0] = ctrl0
	s.args[1] = ctrl1
	s.args[2] = end
	return s
}

// Op returns a clip operation that approximates stroke.
func (s Stroke) Op(ops *op.Ops) clip.Op {
	if len(s.Path.Segments) == 0 {
		return clip.Op{}
	}

	// Approximate and output path data.
	var outline clip.Path
	outline.Begin(ops)
	quads := strokePathCommands(s)
	pen := f32.Pt(0, 0)
	for _, quad := range quads {
		q := quad.Quad
		if q.From != pen {
			pen = q.From
			outline.MoveTo(pen)
		}
		outline.QuadTo(q.Ctrl, q.To)
		pen = q.To
	}
	return clip.Outline{Path: outline.End()}.Op()
}
