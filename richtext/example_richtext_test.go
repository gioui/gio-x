package richtext_test

import (
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/gesture"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
)

func Example() {
	var (
		fonts = gofont.Collection()
		th    = material.NewTheme()
		black = color.NRGBA{A: 255}
		green = color.NRGBA{G: 170, A: 255}
		blue  = color.NRGBA{B: 170, A: 255}
		red   = color.NRGBA{R: 170, A: 255}
	)
	th.Shaper = text.NewShaper(text.WithCollection(fonts))
	go func() {
		w := new(app.Window)

		// allocate persistent state for interactive text. This
		// needs to be persisted across frames.
		var state richtext.InteractiveText

		interactColors := []color.NRGBA{black, green, blue, red}
		interactColorIndex := 0

		var ops op.Ops
		for {
			e := w.Event()
			switch e := e.(type) {
			case app.DestroyEvent:
				panic(e.Err)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				// define the text that you want to present. This can be persisted
				// across frames, recomputed every frame, or modified in any way between
				// frames.
				var spans []richtext.SpanStyle = []richtext.SpanStyle{
					{
						Content: "Hello ",
						Color:   black,
						Size:    unit.Sp(24),
						Font:    fonts[0].Font,
					},
					{
						Content: "in ",
						Color:   green,
						Size:    unit.Sp(36),
						Font:    fonts[0].Font,
					},
					{
						Content: "rich ",
						Color:   blue,
						Size:    unit.Sp(30),
						Font:    fonts[0].Font,
					},
					{
						Content: "text\n",
						Color:   red,
						Size:    unit.Sp(40),
						Font:    fonts[0].Font,
					},
					{
						Content:     "Interact with me!",
						Color:       interactColors[interactColorIndex%len(interactColors)],
						Size:        unit.Sp(40),
						Font:        fonts[0].Font,
						Interactive: true,
					},
				}

				// process any interactions with the text since the last frame.
				for {
					span, event, ok := state.Update(gtx)
					if !ok {
						break
					}
					content, _ := span.Content()
					switch event.Type {
					case richtext.Click:
						log.Println(event.ClickData.Kind)
						if event.ClickData.Kind == gesture.KindClick {
							interactColorIndex++
							gtx.Execute(op.InvalidateCmd{})
						}
					case richtext.Hover:
						w.Option(app.Title("Hovered: " + content))
					case richtext.Unhover:
						w.Option(app.Title("Unhovered: " + content))
					case richtext.LongPress:
						w.Option(app.Title("Long-pressed: " + content))
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
