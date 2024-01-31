package outlay

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

func TestGridLockedRows(t *testing.T) {
	var grid Grid
	var ops op.Ops
	var gtx layout.Context = layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      (&input.Router{}).Source(),
		Now:         time.Time{},
		Locale:      system.Locale{},
		Ops:         &ops,
	}

	highestX := 0
	highestY := 0

	dimensioner := func(axis layout.Axis, index, constraint int) int {
		return 10
	}
	layoutCell := func(gtx layout.Context, x, y int) layout.Dimensions {
		if x > highestX {
			highestX = x
		}
		if y > highestY {
			highestY = y
		}
		return layout.Dimensions{Size: image.Pt(10, 10)}
	}

	grid.Layout(gtx, 10, 10, dimensioner, layoutCell)

	if highestX != 9 {
		t.Errorf("expected highest X index laid out to be %d, got %d", 9, highestX)
	}
	if highestY != 9 {
		t.Errorf("expected highest Y index laid out to be %d, got %d", 9, highestY)
	}

	highestX = 0
	highestY = 0
	grid.LockedRows = 3
	grid.Layout(gtx, 10, 10, dimensioner, layoutCell)

	if highestX != 9 {
		t.Errorf("expected highest X index laid out to be %d, got %d", 9, highestX)
	}
	if highestY != 9 {
		t.Errorf("expected highest Y index laid out to be %d, got %d", 9, highestY)
	}
}

func TestGridSize(t *testing.T) {
	var grid Grid
	var ops op.Ops
	var gtx layout.Context = layout.Context{
		Constraints: layout.Constraints{
			Min: image.Pt(10, 10),
			Max: image.Pt(1000, 1000),
		},
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source: (&input.Router{}).Source(),
		Now:    time.Time{},
		Locale: system.Locale{},
		Ops:    &ops,
	}

	dimensioner := func(axis layout.Axis, index, constraint int) int {
		return 10
	}
	layoutCell := func(gtx layout.Context, x, y int) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(10, 10)}
	}

	// Ensure the returned size is less than the maximum.
	dims := grid.Layout(gtx, 10, 10, dimensioner, layoutCell)
	expected := (layout.Dimensions{Size: image.Pt(100, 100)})
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}

	// Ensure returned size respects maximum.
	gtx.Constraints.Max = image.Pt(50, 50)
	dims = grid.Layout(gtx, 10, 10, dimensioner, layoutCell)
	expected = (layout.Dimensions{Size: image.Pt(50, 50)})
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}

	// Ensure returned size respects minimum.
	gtx.Constraints.Min = image.Pt(500, 500)
	gtx.Constraints.Max = image.Pt(1000, 1000)
	dims = grid.Layout(gtx, 10, 10, dimensioner, layoutCell)
	expected = (layout.Dimensions{Size: image.Pt(500, 500)})
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}
}

func TestGridPointerEvents(t *testing.T) {
	var grid Grid
	var ops op.Ops
	router := &input.Router{}
	var gtx layout.Context = layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      router.Source(),
		Now:         time.Time{},
		Locale:      system.Locale{},
		Ops:         &ops,
	}

	sideSize := 100

	dimensioner := func(axis layout.Axis, index, constraint int) int {
		return sideSize
	}
	layoutCell := func(gtx layout.Context, x, y int) layout.Dimensions {
		defer clip.Rect{Max: image.Pt(sideSize, sideSize)}.Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, t)
		return layout.Dimensions{Size: image.Pt(sideSize, sideSize)}
	}

	// Lay out the grid to establish its input handlers.
	grid.Layout(gtx, 1, 1, dimensioner, layoutCell)
	router.Frame(gtx.Ops)

	// Drain the initial cancel event:
	_, _ = router.Event(pointer.Filter{
		Target: t,
		Kinds:  pointer.Press,
	})

	// Queue up a press.
	press := pointer.Event{
		Position: f32.Point{
			X: 50,
			Y: 50,
		},
		Kind: pointer.Press,
	}
	router.Queue(press)

	event, ok := router.Event(pointer.Filter{
		Target: t,
		Kinds:  pointer.Press,
	})
	if !ok {
		t.Errorf("expected an event, got none")
	} else if event.(pointer.Event).Kind != press.Kind {
		t.Errorf("expected %#+v, got %#+v", press, event)
	}
}
