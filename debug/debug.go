// Package debug provides tools for layout debugging.
package debug

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"sync"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"golang.org/x/image/font/gofont/gomono"
)

var (
	mapLock    sync.RWMutex
	stateMap   map[any]*ConstraintEditor
	active     *ConstraintEditor
	shaperLock sync.Mutex
	shaper     *text.Shaper
)

func init() {
	mapLock.Lock()
	defer mapLock.Unlock()
	stateMap = make(map[any]*ConstraintEditor)
	shaperLock.Lock()
	defer shaperLock.Unlock()
	coll, _ := opentype.ParseCollection(gomono.TTF)
	shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(coll))
}

func getTag(tag any) *ConstraintEditor {
	var state *ConstraintEditor
	func() {
		mapLock.RLock()
		defer mapLock.RUnlock()
		state = stateMap[tag]
	}()
	if state == nil {
		mapLock.Lock()
		defer mapLock.Unlock()
		state = &ConstraintEditor{}
		stateMap[tag] = state
	}
	return state
}

func getActiveEditor() *ConstraintEditor {
	mapLock.RLock()
	defer mapLock.RUnlock()
	return active
}

func setActiveEditor(editor *ConstraintEditor) {
	mapLock.Lock()
	defer mapLock.Unlock()
	active = editor
}

// Wrap wraps w with a [debug.ConstraintEditor]. The state for the constraint
// editor is stored automatically using the unique tag provided. Note: the state
// will never be deleted.
func Wrap(tag any, w layout.Widget) layout.Widget {
	return getTag(tag).Wrap(w)
}

// Layout wraps w with a [debug.ConstraintEditor]. The state for the constraint
// editor is stored automatically using the unique tag provided. Note: the state
// will never be deleted.
func Layout(gtx layout.Context, tag any, w layout.Widget) layout.Dimensions {
	return getTag(tag).Layout(gtx, w)
}

// dragBox holds state for a draggable, hoverable rectangle with drag that can
// accumulate.
type dragBox struct {
	committedDrag    image.Point
	activeDrag       image.Point
	activeDragOrigin image.Point
	drag             gesture.Drag
	hover            gesture.Hover
	hovering         bool
}

// Add inserts the dragBox's input operations into the ops list.
func (d *dragBox) Add(ops *op.Ops) {
	d.drag.Add(ops)
	d.hover.Add(ops)
	pointer.CursorSouthEastResize.Add(ops)
}

// Update processes events from the queue using the given metric and updates the
// drag position.
func (d *dragBox) Update(metric unit.Metric, queue input.Source) {
	for {
		ev, ok := d.drag.Update(metric, queue, gesture.Both)
		if !ok {
			break
		}
		switch ev.Kind {
		case pointer.Press:
			d.activeDragOrigin = ev.Position.Round()
		case pointer.Cancel, pointer.Release:
			d.activeDragOrigin = image.Point{}
			d.committedDrag = d.committedDrag.Add(d.activeDrag)
			d.activeDrag = image.Point{}
		case pointer.Drag:
			d.activeDrag = d.activeDragOrigin.Sub(ev.Position.Round())
		}
	}
	d.hovering = d.hover.Update(queue)
}

// CurrentDrag returns the current accumulated drag (both drag from previous events
// and any in-progress drag gesture).
func (d *dragBox) CurrentDrag() image.Point {
	return d.committedDrag.Add(d.activeDrag)
}

// Active returns whether the user is hovering or interacting with the dragBox.
// It assumes Update() has already been invoked for the current frame.
func (d *dragBox) Active() bool {
	return d.drag.Dragging() || d.drag.Pressed() || d.hovering
}

// Reset clears all accumulated drag.
func (d *dragBox) Reset() {
	d.committedDrag = image.Point{}
}

// ConstraintEditor provides controls to edit layout constraints live.
type ConstraintEditor struct {
	maxBox  dragBox
	minBox  dragBox
	click   gesture.Click
	focused bool

	// LineWidth is the width of debug overlay lines like those outlining the constraints
	// and widget size.
	LineWidth unit.Dp
	// MinSize is the side length of the smallest the editor is allowed to go. If the editor
	// makes the constraints smaller than this, it will reset itself. If the constraints are
	// already smaller than this, the editor will not display itself.
	MinSize unit.Dp
	// TextSize determines the size of the on-screen contextual help text.
	TextSize unit.Sp

	MinColor, MaxColor, SizeColor, SurfaceColor color.NRGBA
}

func outline(ops *op.Ops, width int, area image.Point) clip.PathSpec {
	widthF := float32(width)
	var builder clip.Path
	builder.Begin(ops)
	builder.LineTo(f32.Pt(float32(area.X), 0))
	builder.LineTo(f32.Pt(float32(area.X), float32(area.Y)))
	builder.LineTo(f32.Pt(0, float32(area.Y)))
	builder.LineTo(f32.Pt(0, 0))
	if area.X > width && area.Y > width {
		builder.MoveTo(f32.Pt(widthF, widthF))
		builder.LineTo(f32.Pt(widthF, float32(area.Y)-widthF))
		builder.LineTo(f32.Pt(float32(area.X)-widthF, float32(area.Y)-widthF))
		builder.LineTo(f32.Pt(float32(area.X)-widthF, widthF))
		builder.LineTo(f32.Pt(widthF, widthF))
	}
	builder.Close()
	return builder.End()
}

// Wrap returns a new layout function with the ConstraintEditor wrapping the provided widget.
// This can make inserting the editor more ergonomic than the Layout signature.
func (c *ConstraintEditor) Wrap(w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return c.Layout(gtx, w)
	}
}

// rgb converts a string of the form "#abcdef" or "#abcdef01" into an NRGBA color.
// If the hex does not provide alpha, max alpha is assumed.
func rgb(s string) color.NRGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s)%2 != 0 || len(s) > 8 {
		panic(fmt.Errorf("invalid color #%s", s))
	}
	r, err := strconv.ParseUint(s[:2], 16, 8)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		panic(err)
	}
	a := uint64(255)
	if len(s) > 6 {
		a, err = strconv.ParseUint(s[6:8], 16, 8)
		if err != nil {
			panic(err)
		}
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

func (c *ConstraintEditor) init() {
	if c.LineWidth == 0 {
		c.LineWidth = 1
	}
	if c.MinSize == 0 {
		c.MinSize = 5
	}
	if c.TextSize == 0 {
		c.TextSize = 12
	}
	if c.MinColor == (color.NRGBA{}) {
		c.MinColor = rgb("#c077f9")
	}
	if c.MaxColor == (color.NRGBA{}) {
		c.MaxColor = rgb("#f95f98")
	}
	if c.SizeColor == (color.NRGBA{}) {
		c.SizeColor = rgb("#1fbd51")
	}
	if c.SurfaceColor == (color.NRGBA{}) {
		// TODO: find better color for this.
		c.SurfaceColor = color.NRGBA{R: 1, G: 1, B: 1, A: 0}
	}
}

func record(gtx layout.Context, w layout.Widget) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	return macro.Stop(), dims
}

func colorMaterial(ops *op.Ops, col color.NRGBA) op.CallOp {
	macro := op.Record(ops)
	paint.ColorOp{Color: col}.Add(ops)
	return macro.Stop()
}

func labelOp(gtx layout.Context, sz unit.Sp, col color.NRGBA, str string) (op.CallOp, layout.Dimensions) {
	gtx.Constraints.Min = image.Point{}
	return record(gtx, func(gtx layout.Context) layout.Dimensions {
		shaperLock.Lock()
		defer shaperLock.Unlock()
		return widget.Label{
			MaxLines: 1,
		}.Layout(gtx, shaper, font.Font{}, sz, str, colorMaterial(gtx.Ops, col))
	})
}

func recorded(call op.CallOp, dims layout.Dimensions) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		call.Add(gtx.Ops)
		return dims
	}
}

// Layout the constraint editor to debug the layout of w.
func (c *ConstraintEditor) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	c.init()

	originalConstraints := gtx.Constraints
	c.maxBox.Update(gtx.Metric, gtx.Source)
	c.minBox.Update(gtx.Metric, gtx.Source)

	for {
		event, ok := c.click.Update(gtx.Source)
		if !ok {
			break
		}
		switch event.Kind {
		case gesture.KindClick:
			gtx.Execute(key.FocusCmd{
				Tag: c,
			})
		}
	}
	for {
		event, ok := gtx.Source.Event(
			key.FocusFilter{
				Target: c,
			},
			key.Filter{
				Focus:    c,
				Optional: key.ModShift,
				Name:     "M",
			},
			key.Filter{
				Focus: c,
				Name:  "R",
			},
			key.Filter{
				Focus: c,
				Name:  key.NameEscape,
			},
		)
		if !ok {
			break
		}
		switch event := event.(type) {
		case key.FocusEvent:
			c.focused = event.Focus
		case key.Event:
			if event.State != key.Release {
				continue
			}
			switch event.Name {
			case key.NameEscape:
				gtx.Execute(key.FocusCmd{Tag: nil})
			case "R":
				c.maxBox.Reset()
				c.minBox.Reset()
			case "M":
				if event.Modifiers.Contain(key.ModShift) {
					c.minBox.committedDrag = originalConstraints.Min.Sub(gtx.Constraints.Max)
				} else {
					c.maxBox.committedDrag = originalConstraints.Max.Sub(gtx.Constraints.Min)
				}
			}
		}
	}
	active := c.focused

	gtx.Constraints = gtx.Constraints.SubMax(c.maxBox.CurrentDrag())
	gtx.Constraints = gtx.Constraints.AddMin(image.Point{}.Sub(c.minBox.CurrentDrag()))
	if minSize := gtx.Dp(c.MinSize); gtx.Constraints.Max.X < minSize || gtx.Constraints.Max.Y < minSize {
		gtx.Constraints = originalConstraints
		c.maxBox.Reset()
		c.minBox.Reset()
	}
	dims := w(gtx)
	lineWidth := gtx.Dp(c.LineWidth)
	sizeSpec := outline(gtx.Ops, lineWidth, dims.Size)
	// Display the static widget size.
	paint.FillShape(gtx.Ops, c.SizeColor, clip.Outline{Path: sizeSpec}.Op())
	sizeClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	if active || c.click.Hovered() {
		sizeFill := c.SizeColor
		sizeFill.A = 50
		paint.FillShape(gtx.Ops, sizeFill, clip.Rect{Max: dims.Size}.Op())
	}
	c.click.Add(gtx.Ops)
	event.Op(gtx.Ops, c)
	sizeClip.Pop()

	if active {
		minSpec := outline(gtx.Ops, lineWidth, gtx.Constraints.Min)
		maxSpec := outline(gtx.Ops, lineWidth, gtx.Constraints.Max)
		// Display textual overlays.
		minText := fmt.Sprintf("(%d,%d) Min", gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
		minOp, minDims := labelOp(gtx, c.TextSize, c.MinColor, minText)
		maxText := fmt.Sprintf("(%d,%d) Max", gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
		maxOp, maxDims := labelOp(gtx, c.TextSize, c.MaxColor, maxText)
		szText := fmt.Sprintf("(%d,%d) Size", dims.Size.X, dims.Size.Y)
		szOp, szDims := labelOp(gtx, c.TextSize, c.SizeColor, szText)
		rec := op.Record(gtx.Ops)

		flexAxis := layout.Vertical
		if minDims.Size.Y+maxDims.Size.Y+szDims.Size.Y > gtx.Constraints.Max.Y {
			flexAxis = layout.Horizontal
		}
		layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				paint.FillShape(gtx.Ops, c.SurfaceColor, clip.Rect{Max: gtx.Constraints.Min}.Op())
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: flexAxis,
				}.Layout(gtx,
					layout.Rigid(recorded(minOp, minDims)),
					layout.Rigid(recorded(maxOp, maxDims)),
					layout.Rigid(recorded(szOp, szDims)),
				)
			}),
		)
		// Display the interactive max constraint controls.
		paint.FillShape(gtx.Ops, c.MaxColor, clip.Outline{Path: maxSpec}.Op())
		if c.maxBox.Active() {
			maxFill := c.MaxColor
			maxFill.A = 50
			paint.FillShape(gtx.Ops, maxFill, clip.Rect{Max: gtx.Constraints.Max}.Op())
		}

		maxDragArea := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		c.maxBox.Add(gtx.Ops)
		maxDragArea.Pop()

		// Display the interactive min constraint controls.
		paint.FillShape(gtx.Ops, c.MinColor, clip.Outline{Path: minSpec}.Op())
		if c.minBox.Active() {
			minFill := c.MinColor
			minFill.A = 50
			paint.FillShape(gtx.Ops, minFill, clip.Rect{Max: gtx.Constraints.Min}.Op())
		}

		minDragArea := clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops)
		c.minBox.Add(gtx.Ops)
		minDragArea.Pop()

		op.Defer(gtx.Ops, rec.Stop())
	}

	return dims
}
