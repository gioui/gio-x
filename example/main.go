package main

import (
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/outlay"
	"git.sr.ht/~whereswaldon/outlay/example/playing"
	xwidget "git.sr.ht/~whereswaldon/outlay/example/widget"
	"git.sr.ht/~whereswaldon/outlay/example/widget/boring"
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

func genCards(th *material.Theme) []boring.HoverCard {
	cards := []boring.HoverCard{}
	max := 10
	deck := playing.Deck()
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	for i := 0; i < max; i++ {
		cards = append(cards, boring.HoverCard{
			CardStyle: boring.CardStyle{
				Card:   deck[i],
				Theme:  th,
				Height: unit.Dp(200),
			},
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
