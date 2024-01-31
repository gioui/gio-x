package styledtext

import (
	"image"
	"testing"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
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
			gtx := app.NewContext(&ops, app.FrameEvent{
				Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
				Size:   tc.space,
			})

			txt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})
		})
	}
}

// TestStyledtextNewlines ensures that newlines create appropriate gaps between text.
func TestStyledtextNewlines(t *testing.T) {
	gtx := app.NewContext(new(op.Ops), app.FrameEvent{
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Size:   image.Point{X: 40, Y: 1000},
	})
	gtx.Constraints.Min = image.Point{}
	shaper := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))

	singleLineTxt := Text(shaper, SpanStyle{Size: 12, Content: "a"})
	singleLineDims := singleLineTxt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})

	type testcase struct {
		name          string
		spans         []SpanStyle
		expectedLines int
	}
	for _, tc := range []testcase{
		{
			name:          "double newline between simple letters",
			expectedLines: 3,
			spans: []SpanStyle{
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
					Size:    16,
					Content: "a",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    16,
					Content: "\n\n",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    16,
					Content: "b",
				},
			},
		},
		{
			name:          "double newline after a too-long word",
			expectedLines: 3,
			spans: []SpanStyle{
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Bold},
					Size:    16,
					Content: "mmmmm a",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    16,
					Content: "\n\n",
				},
				{
					Font:    font.Font{Typeface: "Go", Style: font.Regular, Weight: font.Normal},
					Size:    16,
					Content: "b",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			txt := Text(shaper, tc.spans...)
			txtDims := txt.Layout(gtx, func(gtx layout.Context, idx int, dims layout.Dimensions) {})

			if expectedMinY := int((float32(tc.expectedLines) - .5) * float32(singleLineDims.Size.Y)); txtDims.Size.Y <= expectedMinY {
				t.Errorf("expected double newline to create %d lines, dimensions too small", tc.expectedLines)
				t.Logf("expected > %d, got %d (single line height is %d)", expectedMinY, txtDims.Size.Y, singleLineDims.Size.Y)
			}
			if expectedMaxY := int((float32(tc.expectedLines) + .5) * float32(singleLineDims.Size.Y)); txtDims.Size.Y <= expectedMaxY {
				t.Errorf("expected double newline to create %d lines, dimensions too large", tc.expectedLines)
				t.Logf("expected < %d, got %d (single line height is %d)", expectedMaxY, txtDims.Size.Y, singleLineDims.Size.Y)
			}
		})
	}
}
