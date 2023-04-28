// SPDX-License-Identifier: Unlicense OR MIT

// Package stroke converts complex strokes to gioui.org/op/clip operations.
package stroke

import (
	"math"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/andybalholm/stroke"
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
	segOpArcTo
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

	// TriangularCap caps stroked paths with a triangular cap, joining the
	// right-hand and left-hand sides of a stroked path with a triangle
	// with height half of the stroked path's width.
	TriangularCap
)

// StrokeJoin describes how stroked paths are collated.
type StrokeJoin uint8

const (
	// RoundJoin joins path segments with a round segment.
	RoundJoin StrokeJoin = iota

	// BevelJoin joins path segments with sharp bevels.
	BevelJoin

	// MiterJoin joins path segments with a sharp corner.
	// It falls back to a bevel join if the miter limit is exceeded.
	MiterJoin
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

func ArcTo(center f32.Point, angle float32) Segment {
	s := Segment{
		op: segOpArcTo,
	}
	s.args[0] = center
	s.args[1].X = angle
	return s
}

// Op returns a clip operation that approximates stroke.
func (s Stroke) Op(ops *op.Ops) clip.Op {
	if len(s.Path.Segments) == 0 {
		return clip.Op{}
	}

	// Use the stroke package to find the outline of the stroke.
	var path [][]stroke.Segment
	var contour []stroke.Segment
	var pen f32.Point

	for _, seg := range s.Path.Segments {
		switch seg.op {
		case segOpMoveTo:
			if len(contour) > 0 {
				path = append(path, contour)
				contour = nil
			}
			pen = seg.args[0]
		case segOpLineTo:
			contour = append(contour, stroke.LinearSegment(stroke.Point(pen), stroke.Point(seg.args[0])))
			pen = seg.args[0]
		case segOpQuadTo:
			contour = append(contour, stroke.QuadraticSegment(stroke.Point(pen), stroke.Point(seg.args[0]), stroke.Point(seg.args[1])))
			pen = seg.args[1]
		case segOpCubeTo:
			contour = append(contour, stroke.Segment{stroke.Point(pen), stroke.Point(seg.args[0]), stroke.Point(seg.args[1]), stroke.Point(seg.args[2])})
			pen = seg.args[2]
		case segOpArcTo:
			var (
				start  = stroke.Point(pen)
				center = stroke.Point(seg.args[0])
				angle  = seg.args[1].X
			)
			switch {
			case absF32(angle) > math.Pi:
				contour = stroke.AppendArc(contour, start, center, angle)
				pen = f32.Point(contour[len(contour)-1].End)
			default:
				out := stroke.ArcSegment(start, center, angle)
				contour = append(contour, out)
				pen = f32.Point(out.End)
			}
		}
	}
	if len(contour) > 0 {
		path = append(path, contour)
	}

	if len(s.Dashes.Dashes) > 0 {
		path = stroke.Dash(path, s.Dashes.Dashes, s.Dashes.Phase)
	}

	var opt stroke.Options
	opt.Width = s.Width
	opt.MiterLimit = s.Miter
	switch s.Cap {
	case RoundCap:
		opt.Cap = stroke.RoundCap
	case SquareCap:
		opt.Cap = stroke.SquareCap
	case FlatCap:
		opt.Cap = stroke.FlatCap
	case TriangularCap:
		opt.Cap = stroke.TriangularCap
	}
	switch s.Join {
	case RoundJoin:
		opt.Join = stroke.RoundJoin
	case BevelJoin:
		opt.Join = stroke.BevelJoin
	case MiterJoin:
		opt.Join = stroke.MiterJoin
	}

	stroked := stroke.Stroke(path, opt)

	// Output path data.
	var outline clip.Path
	outline.Begin(ops)
	for _, contour := range stroked {
		for i, seg := range contour {
			if i == 0 {
				outline.MoveTo(f32.Point(seg.Start))
				pen = f32.Point(seg.Start)
			}
			if pen != f32.Point(seg.Start) {
				outline.LineTo(f32.Point(seg.Start))
			}
			outline.CubeTo(f32.Point(seg.CP1), f32.Point(seg.CP2), f32.Point(seg.End))
			pen = f32.Point(seg.End)
		}
	}

	return clip.Outline{Path: outline.End()}.Op()
}

func absF32(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}
