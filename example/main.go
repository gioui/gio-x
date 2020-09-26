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
	"gioui.org/unit"
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

type Suit uint8
type Rank uint8
type Color bool

const (
	Spades Suit = iota
	Clubs
	Hearts
	Diamonds
)

const (
	Ace Rank = iota
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

const (
	Red   Color = true
	Black Color = false
)

type Card struct {
	Suit
	Rank
}

func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Two:
		return "2"
	case Three:
		return "3"
	case Four:
		return "4"
	case Five:
		return "5"
	case Six:
		return "6"
	case Seven:
		return "7"
	case Eight:
		return "8"
	case Nine:
		return "9"
	case Ten:
		return "10"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return "?"
	}
}

func (s Suit) String() string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	default:
		return "?"
	}
}

func (s Suit) Color() Color {
	switch s {
	case Spades, Clubs:
		return Black
	case Hearts, Diamonds:
		return Red
	default:
		return Black
	}
}

type (
	C = layout.Context
	D = layout.Dimensions
)

type CardState struct {
	*material.Theme
	Card
	Height   unit.Value
	hovering bool
}

func (c *CardState) Hovering(gtx C) bool {
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

const borderWidth = 0.03

func (c *CardState) layoutFace(gtx C) D {
	gtx.Constraints.Max.Y = gtx.Px(c.Height)
	gtx.Constraints.Max.X = int(float32(gtx.Constraints.Max.Y) / cardHeightToWidth)
	outerRadius := float32(gtx.Constraints.Max.X) * cardRadiusToWidth
	innerRadius := (1 - borderWidth) * outerRadius

	borderWidth := c.Height.Scale(borderWidth)
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return Rect{
				Color: color.RGBA{
					R: 255,
					A: 255,
				},
				Size:  layout.FPt(gtx.Constraints.Max),
				Radii: outerRadius,
			}.Layout(gtx)
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(borderWidth).Layout(gtx, func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return Rect{
							Color: color.RGBA{
								R: 255,
								G: 255,
								B: 255,
								A: 255,
							},
							Size:  layout.FPt(gtx.Constraints.Max),
							Radii: innerRadius,
						}.Layout(gtx)
					}),
					layout.Stacked(func(gtx C) D {
						return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
							defer op.Push(gtx.Ops).Pop()
							origin := f32.Point{
								X: float32(gtx.Constraints.Max.X / 2),
								Y: float32(gtx.Constraints.Max.Y / 2),
							}
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

func (c *CardState) layoutCorner(gtx layout.Context) layout.Dimensions {
	var col color.RGBA
	if c.Suit.Color() == Red {
		col = color.RGBA{R: 255, A: 255}
	} else {
		col = color.RGBA{A: 255}
	}
	return layout.NW.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := material.Body1(c.Theme, c.Rank.String())
					label.Color = col
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					label := material.Body1(c.Theme, c.Suit.String())
					label.Color = col
					return label.Layout(gtx)
				}),
			)
		})
	})
}

func (c *CardState) Layout(gtx C) D {
	defer op.Push(gtx.Ops).Pop()

	dims := c.layoutFace(gtx)
	pointer.Rect(image.Rectangle{Max: dims.Size}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   c,
		Types: pointer.Enter | pointer.Leave,
	}.Add(gtx.Ops)
	return dims
}

const cardHeightToWidth = 14.0 / 9.0
const cardRadiusToWidth = 1.0 / 9.0

func genCards(th *material.Theme) []CardState {
	cards := []CardState{}
	max := 10
	for i := 0; i < max; i++ {
		cards = append(cards, CardState{
			Theme:  th,
			Height: unit.Dp(200),
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
