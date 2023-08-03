package styledtext

import (
	"image"
	"testing"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
)

// TestStyledtextRegressions checks for known regressions that have made styledtext hang in the
// past.
func TestStyledtextRegressions(t *testing.T) {
	type testcase struct {
		name  string
		spans []SpanStyle
		space image.Point
	}
	for _, tc := range []testcase{
		{
			name: "single newline in a span",
			spans: []SpanStyle{
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
					Size:    12,
					Content: "Label: ",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    12,
					Content: "select",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    12,
					Content: "\n",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
					Size:    12,
					Content: "Start: ",
				},
			},
			space: image.Point{X: 10, Y: 100},
		},
		{
			name: "paragraphs separated by double newline",
			spans: []SpanStyle{
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
					Size:    12,
					Content: "hi",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    12,
					Content: "\n\n",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    12,
					Content: "there",
				},
			},
			space: image.Point{X: 100, Y: 100},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			txt := Text(text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection())), tc.spans...)
			var ops op.Ops
			gtx := layout.NewContext(&ops, system.FrameEvent{
				Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
				Size:   tc.space,
			})

			txt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})
		})
	}
}
