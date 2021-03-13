/*
This file is derived from work by Egon Elbre in his gio experiments
repository available here:

https://github.com/egonelbre/expgio/tree/master/box-shadows

He generously licensed it under the Unlicense, and thus is is
reproduced here under the same terms.
*/
package component

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type SurfaceLayoutStyle struct {
	DarkMode     bool
	Background   color.NRGBA
	CornerRadius unit.Value
	Elevation    unit.Value
}

func (s *SurfaceLayoutStyle) Layout(gtx layout.Context) layout.Dimensions {
	sz := gtx.Constraints.Min
	rr := float32(gtx.Px(s.CornerRadius))

	r := f32.Rect(0, 0, float32(sz.X), float32(sz.Y))
	s.layoutShadow(gtx, r, rr)
	clip.UniformRRect(r, rr).Add(gtx.Ops)

	background := s.Background
	if s.DarkMode {
		// figure out dark mode blending and support after more important
		// functionality is done.
		// p := darkBlend(s.Elevation.V)
		// background = f32color.LinearFromSRGB(background).Lighten(p).SRGB()
	}
	paint.Fill(gtx.Ops, background)

	return layout.Dimensions{Size: sz}
}

func darkBlend(x float32) float32 {
	if x <= 0 {
		return 0
	}
	p := 15.77125 - 15.77125/float32(math.Pow(2, float64(x)/3.438155))
	if p <= 0 {
		return 0
	} else if p > 16 {
		return 16 * 0.01
	}
	return p * 0.01
}

func (s *SurfaceLayoutStyle) layoutShadow(gtx layout.Context, r f32.Rectangle, rr float32) {
	if s.Elevation.V <= 0 {
		return
	}

	offset := pxf(gtx.Metric, s.Elevation)

	ambient := r
	gradientBox(gtx.Ops, ambient, rr, offset/2, color.NRGBA{A: 0x10})

	penumbra := r.Add(f32.Pt(0, offset/2))
	gradientBox(gtx.Ops, penumbra, rr, offset, color.NRGBA{A: 0x20})

	umbra := outset(penumbra, -offset/2)
	gradientBox(gtx.Ops, umbra, rr/4, offset/2, color.NRGBA{A: 0x30})
}

func gradientBox(ops *op.Ops, r f32.Rectangle, rr, spread float32, col color.NRGBA) {
	transparent := col
	transparent.A = 0

	// ensure we are aligned to pixel grid
	r = round(r)
	rr = float32(math.Ceil(float64(rr)))
	spread = float32(math.Ceil(float64(spread)))

	// calculate inside and outside boundaries
	inside := imageRect(outset(r, -rr))
	center := imageRect(r)
	outside := imageRect(outset(r, spread))

	radialStop2 := image.Pt(0, int(spread+rr))
	radialOffset1 := rr / (spread + rr)

	corners := []func(image.Rectangle) image.Point{
		topLeft,
		topRight,
		bottomRight,
		bottomLeft,
	}

	for _, corner := range corners {
		func() {
			defer op.Save(ops).Load()
			clipr := image.Rectangle{
				Min: corner(inside),
				Max: corner(outside),
			}.Canon()
			clip.Rect(clipr).Add(ops)
			paint.RadialGradientOp{
				Color1: col, Color2: transparent,
				Stop1:   layout.FPt(corner(inside)),
				Stop2:   layout.FPt(corner(inside).Add(radialStop2)),
				Offset1: radialOffset1,
			}.Add(ops)
			paint.PaintOp{}.Add(ops)
		}()
	}

	// top
	func() {
		defer op.Save(ops).Load()
		clipr := image.Rectangle{
			Min: image.Point{
				X: inside.Min.X,
				Y: outside.Min.Y,
			},
			Max: image.Point{
				X: inside.Max.X,
				Y: center.Min.Y,
			},
		}
		clip.Rect(clipr).Add(ops)
		paint.LinearGradientOp{
			Color1: col, Color2: transparent,
			Stop1: layout.FPt(image.Point{
				X: inside.Min.X,
				Y: center.Min.Y,
			}),
			Stop2: layout.FPt(image.Point{
				X: inside.Min.X,
				Y: outside.Min.Y,
			}),
		}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}()

	// right
	func() {
		defer op.Save(ops).Load()
		clipr := image.Rectangle{
			Min: image.Point{
				X: center.Max.X,
				Y: inside.Min.Y,
			},
			Max: image.Point{
				X: outside.Max.X,
				Y: inside.Max.Y,
			},
		}
		clip.Rect(clipr).Add(ops)
		paint.LinearGradientOp{
			Color1: col, Color2: transparent,
			Stop1: layout.FPt(image.Point{
				X: center.Max.X,
				Y: inside.Min.Y,
			}),
			Stop2: layout.FPt(image.Point{
				X: outside.Max.X,
				Y: inside.Min.Y,
			}),
		}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}()

	// bottom
	func() {
		defer op.Save(ops).Load()
		clipr := image.Rectangle{
			Min: image.Point{
				X: inside.Min.X,
				Y: center.Max.Y,
			},
			Max: image.Point{
				X: inside.Max.X,
				Y: outside.Max.Y,
			},
		}
		clip.Rect(clipr).Add(ops)
		paint.LinearGradientOp{
			Color1: col, Color2: transparent,
			Stop1: layout.FPt(image.Point{
				X: inside.Min.X,
				Y: center.Max.Y,
			}),
			Stop2: layout.FPt(image.Point{
				X: inside.Min.X,
				Y: outside.Max.Y,
			}),
		}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}()

	// left
	func() {
		defer op.Save(ops).Load()
		clipr := image.Rectangle{
			Min: image.Point{
				X: outside.Min.X,
				Y: inside.Min.Y,
			},
			Max: image.Point{
				X: center.Min.X,
				Y: inside.Max.Y,
			},
		}
		clip.Rect(clipr).Add(ops)
		paint.LinearGradientOp{
			Color1: col, Color2: transparent,
			Stop1: layout.FPt(image.Point{
				X: center.Min.X,
				Y: inside.Min.Y,
			}),
			Stop2: layout.FPt(image.Point{
				X: outside.Min.X,
				Y: inside.Min.Y,
			}),
		}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}()

	func() {
		defer op.Save(ops).Load()
		var p clip.Path
		p.Begin(ops)

		inside := layout.FRect(inside)
		center := layout.FRect(center)

		p.MoveTo(inside.Min)
		p.LineTo(f32.Point{X: inside.Min.X, Y: center.Min.Y})
		p.LineTo(f32.Point{X: inside.Max.X, Y: center.Min.Y})
		p.LineTo(f32.Point{X: inside.Max.X, Y: inside.Min.Y})
		p.LineTo(f32.Point{X: center.Max.X, Y: inside.Min.Y})
		p.LineTo(f32.Point{X: center.Max.X, Y: inside.Max.Y})
		p.LineTo(f32.Point{X: inside.Max.X, Y: inside.Max.Y})
		p.LineTo(f32.Point{X: inside.Max.X, Y: center.Max.Y})
		p.LineTo(f32.Point{X: inside.Min.X, Y: center.Max.Y})
		p.LineTo(f32.Point{X: inside.Min.X, Y: inside.Max.Y})
		p.LineTo(f32.Point{X: center.Min.X, Y: inside.Max.Y})
		p.LineTo(f32.Point{X: center.Min.X, Y: inside.Min.Y})
		p.LineTo(inside.Min)

		clip.Outline{Path: p.End()}.Op().Add(ops)
		paint.ColorOp{Color: col}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}()
}

func imageRect(r f32.Rectangle) image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: int(math.Round(float64(r.Min.X))),
			Y: int(math.Round(float64(r.Min.Y))),
		},
		Max: image.Point{
			X: int(math.Round(float64(r.Max.X))),
			Y: int(math.Round(float64(r.Max.Y))),
		},
	}
}

func round(r f32.Rectangle) f32.Rectangle {
	return f32.Rectangle{
		Min: f32.Point{
			X: float32(math.Round(float64(r.Min.X))),
			Y: float32(math.Round(float64(r.Min.Y))),
		},
		Max: f32.Point{
			X: float32(math.Round(float64(r.Max.X))),
			Y: float32(math.Round(float64(r.Max.Y))),
		},
	}
}

func outset(r f32.Rectangle, rr float32) f32.Rectangle {
	r.Min.X -= rr
	r.Min.Y -= rr
	r.Max.X += rr
	r.Max.Y += rr
	return r
}

func pxf(c unit.Metric, v unit.Value) float32 {
	switch v.U {
	case unit.UnitPx:
		return v.V
	case unit.UnitDp:
		s := c.PxPerDp
		if s == 0 {
			s = 1
		}
		return s * v.V
	case unit.UnitSp:
		s := c.PxPerSp
		if s == 0 {
			s = 1
		}
		return s * v.V
	default:
		panic("unknown unit")
	}
}

func topLeft(r image.Rectangle) image.Point     { return r.Min }
func topRight(r image.Rectangle) image.Point    { return image.Point{X: r.Max.X, Y: r.Min.Y} }
func bottomRight(r image.Rectangle) image.Point { return r.Max }
func bottomLeft(r image.Rectangle) image.Point  { return image.Point{X: r.Min.X, Y: r.Max.Y} }
