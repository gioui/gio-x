package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/outlay"
	"git.sr.ht/~whereswaldon/sprig/anim"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type (
	C = layout.Context
	D = layout.Dimensions
)

type Card struct {
	Rect
	hovering bool
}

func (c *Card) Hovering(gtx C) bool {
	start := c.hovering
	for _, ev := range gtx.Events(c) {
		switch ev := ev.(type) {
		case pointer.Event:
			switch ev.Type {
			case pointer.Enter:
				c.hovering = true
			case pointer.Leave:
				c.hovering = false
			case pointer.Cancel:
				c.hovering = false
			}
		}
	}
	if c.hovering != start {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	return c.hovering
}

func (c *Card) Layout(gtx C) D {
	defer op.Push(gtx.Ops).Pop()

	dims := c.Rect.Layout(gtx)
	pointer.Rect(image.Rectangle{Max: dims.Size}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   c,
		Types: pointer.Enter | pointer.Leave,
	}.Add(gtx.Ops)
	return dims
}

func genCards() []Card {
	cardSize := f32.Point{
		X: 270.0,
		Y: 420.0,
	}
	radii := float32(30)
	cards := []Card{}
	max := 10
	step := 255 / (max - 1)
	for i := 0; i < max; i++ {
		cards = append(cards, Card{
			Rect: Rect{
				Size: cardSize,
				Color: color.RGBA{
					A: 255,
					R: 255 - uint8(i*step),
				},
				Radii: radii,
			},
		})
	}
	return cards
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	fan := outlay.Fan{
		Normal: anim.Normal{
			Duration: time.Second,
		},
	}
	numCards := widget.Float{}
	cardChildren := []outlay.FanItem{}
	cards := genCards()
	for i := range cards {
		cardChildren = append(cardChildren, outlay.Item(i == 5, cards[i].Layout))
	}
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			for i := range cards {
				cardChildren[i].Elevate = cards[i].Hovering(gtx)
			}
			visibleCards := int(math.Round(float64(numCards.Value*float32(len(cardChildren)-1)))) + 1
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return material.Body1(th, "1").Layout(gtx)
						}),
						layout.Flexed(1, func(gtx C) D {
							return material.Slider(th, &numCards, 0.0, 1.0).Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return material.Body1(th, "10").Layout(gtx)
						}),
					)
				}),
				layout.Flexed(1, func(gtx C) D {
					return fan.Layout(gtx, cardChildren[:visibleCards]...)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}

// Rect creates a rectangle of the provided background color with
// Dimensions specified by size and a corner radius (on all corners)
// specified by radii.
type Rect struct {
	Color color.RGBA
	Size  f32.Point
	Radii float32
}

// Layout renders the Rect into the provided context
func (r Rect) Layout(gtx C) D {
	return DrawRect(gtx, r.Color, r.Size, r.Radii)
}

// DrawRect creates a rectangle of the provided background color with
// Dimensions specified by size and a corner radius (on all corners)
// specified by radii.
func DrawRect(gtx C, background color.RGBA, size f32.Point, radii float32) D {
	stack := op.Push(gtx.Ops)
	paint.ColorOp{Color: background}.Add(gtx.Ops)
	bounds := f32.Rectangle{Max: size}
	if radii != 0 {
		clip.RRect{
			Rect: bounds,
			NW:   radii,
			NE:   radii,
			SE:   radii,
			SW:   radii,
		}.Add(gtx.Ops)
	}
	paint.PaintOp{Rect: bounds}.Add(gtx.Ops)
	stack.Pop()
	return layout.Dimensions{Size: image.Pt(int(size.X), int(size.Y))}
}
