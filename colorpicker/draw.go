package colorpicker

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
	"image/color"
	"math"
)

const c = 0.55228475 // 4*(sqrt(2)-1)/3

func drawControl(p f32.Point, radius, width float32, gtx layout.Context) {
	width = float32(gtx.Px(unit.Dp(width)))
	radius = float32(gtx.Px(unit.Dp(radius))) - width
	p.X -= radius - width*2
	p.Y -= radius - width*4
	drawCircle(p, radius, width, color.NRGBA{A: 0xff}, gtx)
	p.X += width
	p.Y += width
	drawCircle(p, radius, width, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, gtx)
}

func drawCircle(p f32.Point, r, width float32, col color.NRGBA, gtx layout.Context) {
	w := r * 2
	defer op.Save(gtx.Ops).Load()
	var path clip.Path
	path.Begin(gtx.Ops)
	path.Move(f32.Point{X: p.X, Y: p.Y})
	path.Move(f32.Point{X: w / 4 * 3, Y: r / 2})
	path.Cube(f32.Point{X: 0, Y: r * c}, f32.Point{X: -r + r*c, Y: r}, f32.Point{X: -r, Y: r})    // SE
	path.Cube(f32.Point{X: -r * c, Y: 0}, f32.Point{X: -r, Y: -r + r*c}, f32.Point{X: -r, Y: -r}) // SW
	path.Cube(f32.Point{X: 0, Y: -r * c}, f32.Point{X: r - r*c, Y: -r}, f32.Point{X: r, Y: -r})   // NW
	path.Cube(f32.Point{X: r * c, Y: 0}, f32.Point{X: r, Y: r - r*c}, f32.Point{X: r, Y: r})      // NE
	clip.Stroke{Path: path.End(), Style: clip.StrokeStyle{Width: width}}.Op().Add(gtx.Ops)
	cons := gtx.Constraints
	dr := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: cons.Max.X, Y: cons.Max.Y}}
	clip.Rect(dr).Add(gtx.Ops)
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func drawCheckerboard(gtx layout.Context) {
	w := gtx.Constraints.Max.X
	h := gtx.Constraints.Max.Y
	paint.FillShape(gtx.Ops, white, clip.Rect{Max: gtx.Constraints.Max}.Op())
	size := h / 2
	defer op.Save(gtx.Ops).Load()
	var path clip.Path
	path.Begin(gtx.Ops)
	count := int(math.Ceil(float64(w / size)))
	for i := 0; i < count; i++ {
		offset := 0
		if math.Mod(float64(i), 2) == 0 {
			offset += size
		}
		path.MoveTo(f32.Point{X: float32(i * size), Y: float32(offset)})
		path.Line(f32.Point{X: float32(size)})
		path.Line(f32.Point{Y: float32(size)})
		path.Line(f32.Point{X: float32(-size)})
		path.Line(f32.Point{Y: float32(-size)})
	}
	clip.Outline{Path: path.End()}.Op().Add(gtx.Ops)
	paint.ColorOp{Color: lightgrey}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

var (
	red     = color.NRGBA{R: 255, A: 255}
	yellow  = color.NRGBA{R: 255, G: 255, A: 255}
	green   = color.NRGBA{G: 255, A: 255}
	cyan    = color.NRGBA{G: 255, B: 255, A: 255}
	blue    = color.NRGBA{B: 255, A: 255}
	magenta = color.NRGBA{R: 255, B: 255, A: 255}

	white     = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	lightgrey = color.NRGBA{R: 0xef, G: 0xef, B: 0xef, A: 0xff}
)

var rainbowColors = []color.NRGBA{red, yellow, green, cyan, blue, magenta, red}

func drawRainbow(gtx layout.Context) layout.Dimensions {
	w := gtx.Constraints.Max.X
	h := gtx.Constraints.Max.Y
	stepCount := len(rainbowColors)
	stepWidth := float32(w / (stepCount - 1))
	offsetX := float32(0)
	color1 := rainbowColors[0]
	for _, color2 := range rainbowColors[1:] {
		stack := op.Save(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Point{offsetX, 0},
			Stop2:  f32.Point{offsetX + stepWidth, 0},
			Color1: color1,
			Color2: color2,
		}.Add(gtx.Ops)
		dr := image.Rectangle{Min: image.Point{X: int(offsetX), Y: 0}, Max: image.Point{X: int(offsetX + stepWidth), Y: h}}
		clip.Rect(dr).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Load()
		color1 = color2
		offsetX += stepWidth
	}
	return layout.Dimensions{Size: image.Point{X: w, Y: h}}
}
