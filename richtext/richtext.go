package richtext

import (
	"image"
	"image/color"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/image/math/fixed"
)

// InteractiveSpan holds the persistent state of rich text that can
// be interacted with by the user. It can report clicks, hovers, and
// long-presses on the text.
type InteractiveSpan struct {
	click        gesture.Click
	pressing     bool
	longPressed  bool
	pressStarted time.Time
	content      string
	metadata     map[string]string
}

// Layout adds the pointer input op for this interactive span and updates its
// state. It uses the most recent pointer.AreaOp as its input area.
func (i *InteractiveSpan) Layout(gtx layout.Context) layout.Dimensions {
	i.click.Add(gtx.Ops)
	if i.click.Pressed() && !i.pressing {
		i.pressStarted = gtx.Now
	} else if i.pressing && gtx.Now.Sub(i.pressStarted) > time.Millisecond*250 {
		i.longPressed = true
	}
	return layout.Dimensions{}
}

// Hovered returns whether this span is hovered.
func (i *InteractiveSpan) Hovered() bool {
	return i.click.Hovered()
}

// Events returns click event information for this span.
func (i *InteractiveSpan) Events(q event.Queue) []gesture.ClickEvent {
	return i.click.Events(q)
}

// LongPressed returns whether this span has been long-pressed.
func (i *InteractiveSpan) LongPressed() bool {
	out := i.longPressed
	i.longPressed = false
	return out
}

// Content returns the text content of the interactive span as well as the
// metadata associated with it.
func (i *InteractiveSpan) Content() (string, map[string]string) {
	return i.content, i.metadata
}

// Get looks up a metadata property on the interactive span.
func (i *InteractiveSpan) Get(key string) string {
	return i.metadata[key]
}

// InteractiveText holds persistent state for a block of text containing
// spans that may be interactive.
type InteractiveText struct {
	Spans   []InteractiveSpan
	current int
}

// next returns an InteractiveSpan that hasn't been used since the last
// call to reset().
func (i *InteractiveText) next() *InteractiveSpan {
	if i.current >= len(i.Spans) {
		i.Spans = append(i.Spans, InteractiveSpan{})
	}
	span := &i.Spans[i.current]
	i.current++
	return span
}

// reset moves the internal iteration cursor back the start of the spans,
// allowing them to be reused. This should be called at the start of every
// layout.
func (i *InteractiveText) reset() {
	i.current = 0
}

// Hovered returns the first hovered span in the interactive text.
func (i *InteractiveText) Hovered() *InteractiveSpan {
	for k := range i.Spans {
		span := &i.Spans[k]
		if span.Hovered() {
			return span
		}
	}
	return nil
}

// LongPressed returns the first long-pressed span in the interactive text.
func (i *InteractiveText) LongPressed() *InteractiveSpan {
	for k := range i.Spans {
		span := &i.Spans[k]
		if span.LongPressed() {
			return span
		}
	}
	return nil
}

// Events returns the first span with unprocessed events and the events that
// need processing for it.
func (i *InteractiveText) Events(q event.Queue) (*InteractiveSpan, []gesture.ClickEvent) {
	for k := range i.Spans {
		span := &i.Spans[k]
		if events := span.Events(q); len(events) > 0 {
			return span, events
		}
	}
	return nil, nil
}

// SpanStyle describes the appearance of a span of styled text.
type SpanStyle struct {
	Font        text.Font
	Size        unit.Value
	Color       color.NRGBA
	Content     string
	Interactive bool
	metadata    map[string]string
}

// spanShape describes the text shaping of a single span.
type spanShape struct {
	offset image.Point
	layout text.Layout
	size   image.Point
}

// Set configures a metadata key-value pair on the span that can be
// retrieved if the span is interacted with.
func (ss SpanStyle) Set(key, value string) {
	if ss.metadata == nil {
		ss.metadata = make(map[string]string)
	}
	ss.metadata[key] = value
}

// Layout renders the span using the provided text shaping.
func (ss SpanStyle) Layout(gtx layout.Context, s text.Shaper, shape spanShape) layout.Dimensions {
	stack := op.Save(gtx.Ops)
	paint.ColorOp{Color: ss.Color}.Add(gtx.Ops)
	op.Offset(layout.FPt(shape.offset)).Add(gtx.Ops)
	s.Shape(ss.Font, fixed.I(gtx.Px(ss.Size)), shape.layout).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Load()
	return layout.Dimensions{Size: shape.size}
}

func (ss SpanStyle) DeepCopy() SpanStyle {
	md := make(map[string]string)
	for k, v := range ss.metadata {
		md[k] = v
	}
	out := ss
	out.metadata = md
	return out
}

// TextStyle presents rich text.
type TextStyle struct {
	State  *InteractiveText
	Styles []SpanStyle
}

// Text constructs a TextStyle.
func Text(state *InteractiveText, styles ...SpanStyle) TextStyle {
	return TextStyle{
		State:  state,
		Styles: styles,
	}
}

// Layout renders the TextStyle using the provided text shaper.
func (t TextStyle) Layout(gtx layout.Context, shaper text.Shaper) layout.Dimensions {
	spans := make([]SpanStyle, len(t.Styles))
	copy(spans, t.Styles)
	t.State.reset()

	var (
		lineDims       image.Point
		overallSize    image.Point
		lineShapes     []spanShape
		lineStartIndex int
		state          *InteractiveSpan
	)

	for i := 0; i < len(spans); i++ {
		// grab the next span
		span := spans[i]

		// constrain the width of the line to the remaining space
		maxWidth := gtx.Constraints.Max.X - lineDims.X

		// shape the text of the current span
		lines := shaper.LayoutString(span.Font, fixed.I(gtx.Px(span.Size)), maxWidth, span.Content)

		// grab the first line of the result and compute its dimensions
		firstLine := lines[0]
		spanWidth := firstLine.Width.Ceil()
		spanHeight := (firstLine.Ascent + firstLine.Descent).Ceil()

		// store the text shaping results for the line
		lineShapes = append(lineShapes, spanShape{
			offset: image.Point{X: lineDims.X},
			size:   image.Point{X: spanWidth, Y: spanHeight},
			layout: firstLine.Layout,
		})

		// update the dimensions of the current line
		lineDims.X += spanWidth
		if lineDims.Y < spanHeight {
			lineDims.Y = spanHeight
		}

		// update the width of the overall text
		if overallSize.X < lineDims.X {
			overallSize.X = lineDims.X
		}

		// if we are breaking the current span across lines or we are on the
		// last span, lay out all of the spans for the line.
		if len(lines) > 1 || i == len(spans)-1 {
			for i, shape := range lineShapes {
				// lay out this span
				span = spans[i+lineStartIndex]
				shape.offset.Y = overallSize.Y + lineDims.Y
				span.Layout(gtx, shaper, shape)

				if !span.Interactive {
					continue
				}
				// grab an interactive state and lay it out atop the text.
				// If we still have a state, this line is a continuation of
				// the previous span and we should use the same state.
				if state == nil {
					state = t.State.next()
					state.content = span.Content
					state.metadata = span.metadata
				}
				stack := op.Save(gtx.Ops)
				op.Offset(layout.FPt(shape.offset)).Add(gtx.Ops)
				pointer.Rect(image.Rectangle{Max: shape.size}).Add(gtx.Ops)
				state.Layout(gtx)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				stack.Load()
			}
			// reset line shaping data and update overall vertical dimensions
			lineShapes = lineShapes[:0]
			overallSize.Y += lineDims.Y
		}

		// if the current span breaks across lines
		if len(lines) > 1 {
			// mark where the next line to be laid out starts
			lineStartIndex = i + 1
			lineDims = image.Point{}

			// ensure the spans slice has room for another span
			spans = append(spans, SpanStyle{})
			// shift existing spans further
			for k := len(spans) - 1; k > i+1; k-- {
				spans[k] = spans[k-1]
			}
			// synthesize and insert a new span
			span.Content = span.Content[len(firstLine.Layout.Text):]
			spans[i+1] = span
		} else {
			// indicate that the next span is not a continuation of the current
			// one.
			state = nil
		}
	}
	return layout.Dimensions{Size: overallSize}
}

//TextObjects represents the whole richtext widget
type TextObjects []*TextObject

//TextObject is one of the objects in the richtext widget
type TextObject struct {
	Font      text.Font
	Size      unit.Value
	Color     color.NRGBA
	Content   string
	Clickable bool
	metadata  map[string]string

	origin     *TextObject
	textLayout text.Layout
	xOffset    int
	size       image.Point
	clicked    int
	hovered    bool
}

//Layout prints out the text at specified offset
func (to *TextObject) Layout(gtx layout.Context, s text.Shaper, off image.Point) layout.Dimensions {
	stack := op.Save(gtx.Ops)
	paint.ColorOp{Color: to.Color}.Add(gtx.Ops)
	op.Offset(layout.FPt(off)).Add(gtx.Ops)
	s.Shape(to.Font, fixed.I(gtx.Px(to.Size)), to.textLayout).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Load()
	return layout.Dimensions{Size: to.size}
}

func (to *TextObject) Clicked() bool {
	if to.clicked <= 0 {
		return false
	}
	to.clicked--
	return true
}

func (to *TextObject) Hovered() bool {
	return to.hovered
}

func (to *TextObject) updateClicks(gtx layout.Context) {
	if !to.Clickable {
		return
	}
	for _, e := range gtx.Events(origin(to)) {
		switch p := e.(type) {
		case pointer.Event:
			switch p.Type {
			case pointer.Release:
				to.clicked++
			case pointer.Enter:
				to.hovered = true
			case pointer.Leave, pointer.Cancel:
				to.hovered = false
			}
		}
	}
}

func (to *TextObject) SetMetadata(key, value string) {
	if to.metadata == nil {
		to.metadata = make(map[string]string)
	}
	to.metadata[key] = value
}

func (to *TextObject) GetMetadata(key string) string {
	if to.metadata == nil {
		return ""
	}
	return to.metadata[key]
}

func (to *TextObject) DeepCopy() *TextObject {
	t := *to
	t.metadata = make(map[string]string)
	for k, v := range to.metadata {
		t.metadata[k] = v
	}
	return &t
}

//Layout prints out the rich text widget
func (tos TextObjects) Layout(gtx layout.Context, s text.Shaper) layout.Dimensions {
	var tosDims layout.Dimensions
	var lineDims image.Point
	oi := &objectIterator{objs: tos}
	//TextObjects (or it's parts) on a single line
	var lineObjects []*TextObject

	for to := oi.Next(); to != nil; to = oi.Next() {
		maxWidth := gtx.Constraints.Max.X - lineDims.X
		//layout the string using the maxWidthe remaining for the line
		lines := s.LayoutString(to.Font, fixed.I(gtx.Px(to.Size)), maxWidth, to.Content)
		//using just the first line
		first := lines[0]
		lineWidth := first.Width.Ceil()
		lineHeight := (first.Ascent + first.Descent).Ceil()

		//getting the used text in the line and calculating the dimensions
		to.textLayout = first.Layout
		to.xOffset = lineDims.X
		to.size = image.Point{X: lineWidth, Y: lineHeight}
		lineObjects = append(lineObjects, to)

		//calculating the X and the max of the Y of the line
		lineDims.X += lineWidth
		if lineDims.Y < lineHeight {
			lineDims.Y = lineHeight
		}
		//the width of the whole richtext object
		if tosDims.Size.X < lineDims.X {
			tosDims.Size.X = lineDims.X
		}

		//if we are going to break the line, or we are at the end of the richtext
		if len(lines) > 1 || oi.Empty() {
			//we print the line
			for _, obj := range lineObjects {
				off := image.Point{
					X: obj.xOffset,
					Y: tosDims.Size.Y + lineDims.Y,
				}
				obj.Layout(gtx, s, off)
				if !obj.Clickable {
					continue
				}
				to.updateClicks(gtx)
				stack := op.Save(gtx.Ops)
				clickableOffset := image.Point{X: obj.xOffset, Y: tosDims.Size.Y}
				op.Offset(layout.FPt(clickableOffset)).Add(gtx.Ops)
				pointer.Rect(image.Rectangle{Max: obj.size}).Add(gtx.Ops)
				pointer.InputOp{
					Tag:   origin(obj),
					Types: pointer.Release | pointer.Enter | pointer.Leave,
				}.Add(gtx.Ops)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				stack.Load()
			}
			//we printed these objectss, so we get rid of them
			lineObjects = nil
			tosDims.Size.Y += lineDims.Y
		}
		if len(lines) > 1 {
			//add the rest of the TextObject to be printed on a new line
			//this puts it to the front of the objectIterator
			oi.Add(&TextObject{
				origin:    origin(to),
				Font:      to.Font,
				Size:      to.Size,
				Color:     to.Color,
				Clickable: to.Clickable,
				Content:   to.Content[len(first.Layout.Text):],
			})
			lineDims.X = 0
			lineDims.Y = 0
		}
	}

	return tosDims
}

type objectIterator struct {
	objs []*TextObject
}

func (oi *objectIterator) Add(to *TextObject) {
	oi.objs = append([]*TextObject{to}, oi.objs...)
}

func (oi *objectIterator) Next() *TextObject {
	if len(oi.objs) == 0 {
		return nil
	}
	next := oi.objs[0]
	oi.objs = oi.objs[1:]
	return next
}

func (oi *objectIterator) Empty() bool {
	return len(oi.objs) == 0
}

func (tos TextObjects) Clicked() *TextObject {
	for _, to := range tos {
		if to.Clicked() {
			return to
		}
	}
	return nil
}

func (tos TextObjects) Hovered() *TextObject {
	for _, to := range tos {
		if to.Hovered() {
			return to
		}
	}
	return nil
}

func origin(to *TextObject) *TextObject {
	if to.origin == nil {
		return to
	}
	return to.origin
}
