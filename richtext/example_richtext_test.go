package richtext

import (
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/gesture"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
)

var (
	fonts = gofont.Collection()
	th    = material.NewTheme(fonts)
	black = color.NRGBA{A: 255}
	green = color.NRGBA{G: 170, A: 255}
	blue  = color.NRGBA{B: 170, A: 255}
	red   = color.NRGBA{R: 170, A: 255}
)

func Example() {
	go func() {
		w := app.NewWindow()

		// allocate persistent state for interactive text. This
		// needs to be persisted across frames.
		var state richtext.InteractiveText

		interactColors := []color.NRGBA{black, green, blue, red}
		interactColorIndex := 0

		var ops op.Ops
		for {
			e := <-w.Events()
			switch e := e.(type) {
			case system.DestroyEvent:
				panic(e.Err)
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				// define the text that you want to present. This can be persisted
				// across frames, recomputed every frame, or modified in any way between
				// frames.
				var spans []richtext.SpanStyle = []richtext.SpanStyle{
					{
						Content: "Hello ",
						Color:   black,
						Size:    unit.Dp(24),
						Font:    fonts[0].Font,
					},
					{
						Content: "in ",
						Color:   green,
						Size:    unit.Dp(36),
						Font:    fonts[0].Font,
					},
					{
						Content: "rich ",
						Color:   blue,
						Size:    unit.Dp(30),
						Font:    fonts[0].Font,
					},
					{
						Content: "text\n",
						Color:   red,
						Size:    unit.Dp(40),
						Font:    fonts[0].Font,
					},
					{
						Content:     "Interact with me!",
						Color:       interactColors[interactColorIndex%len(interactColors)],
						Size:        unit.Dp(40),
						Font:        fonts[0].Font,
						Interactive: true,
					},
				}

				// process any interactions with the text since the last frame.
				for span, events := state.Events(); span != nil; span, events = state.Events() {
					for _, event := range events {
						content, _ := span.Content()
						switch event.Type {
						case richtext.Click:
							log.Println(event.ClickData.Type)
							if event.ClickData.Type == gesture.TypeClick {
								interactColorIndex++
								op.InvalidateOp{}.Add(gtx.Ops)
							}
						case richtext.Hover:
							w.Option(app.Title("Hovered: " + content))
						case richtext.LongPress:
							w.Option(app.Title("Long-pressed: " + content))
						}
					}
				}

				// render the rich text into the operation list
				richtext.Text(&state, th.Shaper, spans...).Layout(gtx)

				// render the operation list
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
