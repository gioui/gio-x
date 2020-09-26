package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/outlay"
	"git.sr.ht/~whereswaldon/outlay/example/playing"
	xwidget "git.sr.ht/~whereswaldon/outlay/example/widget"
	"git.sr.ht/~whereswaldon/sprig/anim"
)

type (
	C = layout.Context
	D = layout.Dimensions
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

type CardPalette struct {
	RedSuit, BlackSuit color.RGBA
	Border, Background color.RGBA
}

func (p CardPalette) ColorFor(s playing.Suit) color.RGBA {
	if s.Color() == playing.Red {
		return p.RedSuit
	}
	return p.BlackSuit
}

var DefaultPalette = &CardPalette{
	RedSuit:    color.RGBA{R: 0xa0, B: 0x20, A: 0xff},
	BlackSuit:  color.RGBA{A: 0xff},
	Border:     color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff},
	Background: color.RGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff},
}

type CardStyle struct {
	*xwidget.CardState
	*material.Theme
	playing.Card
	Height unit.Value
	*CardPalette
}

const cardHeightToWidth = 14.0 / 9.0
const cardRadiusToWidth = 1.0 / 16.0
const borderWidth = 0.005

func (c *CardStyle) Palette() *CardPalette {
	if c.CardPalette == nil {
		return DefaultPalette
	}
	return c.CardPalette
}

func (c *CardStyle) layoutFace(gtx C) D {
	gtx.Constraints.Max.Y = gtx.Px(c.Height)
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.Y) / cardHeightToWidth)
	outerRadius := float32(gtx.Constraints.Max.X) * cardRadiusToWidth
	innerRadius := (1 - borderWidth) * outerRadius

	borderWidth := c.Height.Scale(borderWidth)
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return Rect{
				Color: c.Palette().Border,
				Size:  layout.FPt(gtx.Constraints.Max),
				Radii: outerRadius,
			}.Layout(gtx)
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(borderWidth).Layout(gtx, func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return Rect{
							Color: c.Palette().Background,
							Size:  layout.FPt(gtx.Constraints.Max),
							Radii: innerRadius,
						}.Layout(gtx)
					}),
					layout.Stacked(func(gtx C) D {
						return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
							defer op.Push(gtx.Ops).Pop()
							gtx.Constraints.Min = gtx.Constraints.Max
							origin := f32.Point{
								X: float32(gtx.Constraints.Max.X / 2),
								Y: float32(gtx.Constraints.Max.Y / 2),
							}
							layout.Center.Layout(gtx, func(gtx C) D {
								face := material.H1(c.Theme, c.Rank.String())
								face.Color = c.Palette().ColorFor(c.Suit)
								return face.Layout(gtx)
							})
							c.layoutCorner(gtx)
							op.Affine(f32.Affine2D{}.Rotate(origin, math.Pi)).Add(gtx.Ops)
							c.layoutCorner(gtx)

							return D{Size: gtx.Constraints.Max}
						})
					}),
				)
			})
		}),
	)
}

func (c *CardStyle) layoutCorner(gtx layout.Context) layout.Dimensions {
	col := c.Palette().ColorFor(c.Suit)
	return layout.NW.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := material.H6(c.Theme, c.Rank.String())
					label.Color = col
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					label := material.H6(c.Theme, c.Suit.String())
					label.Color = col
					return label.Layout(gtx)
				}),
			)
		})
	})
}

func (c *CardStyle) Layout(gtx C) D {
	dims := c.layoutFace(gtx)
	gtx.Constraints.Max = dims.Size
	c.CardState.Layout(gtx)
	return dims
}

func genCards(th *material.Theme) []CardStyle {
	cards := []CardStyle{}
	max := 10
	deck := playing.Deck()
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	for i := 0; i < max; i++ {
		cards = append(cards, CardStyle{
			Card:      deck[i],
			Theme:     th,
			Height:    unit.Dp(200),
			CardState: &xwidget.CardState{},
		})
	}
	return cards
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	fan := outlay.Fan{
		Normal: anim.Normal{
			Duration: time.Second / 4,
		},
	}
	numCards := widget.Float{}
	cardChildren := []outlay.FanItem{}
	cards := genCards(th)
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
