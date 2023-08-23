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

// TestStyledtextNewlines ensures that newlines create appropriate gaps between text.
func TestStyledtextNewlines(t *testing.T) {
	gtx := layout.NewContext(new(op.Ops), system.FrameEvent{
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Size:   image.Point{X: 1000, Y: 1000},
	})
	gtx.Constraints.Min = image.Point{}
	shaper := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))

	singleLineTxt := Text(shaper, SpanStyle{Size: 12, Content: "a"})
	singleLineDims := singleLineTxt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})

	txt := Text(shaper,
		SpanStyle{
			Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
			Size:    12,
			Content: "a",
		},
		SpanStyle{
			Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
			Size:    12,
			Content: "\n\n",
		},
		SpanStyle{
			Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
			Size:    12,
			Content: "b",
		},
	)
	txtDims := txt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})

	if expectedY := int(2.5 * float32(singleLineDims.Size.Y)); txtDims.Size.Y <= expectedY {
		t.Errorf("expected double newline to create 3 lines, dimensions too small")
		t.Logf("expected > %d, got %d (single line height is %d)", expectedY, txtDims.Size.Y, singleLineDims.Size.Y)
	}
}
