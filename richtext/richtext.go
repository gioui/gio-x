/*
Package richtext provides rendering of text containing multiple fonts, styles, and levels of interactivity.
*/
package richtext

import (
	"image"
	"image/color"
	"time"
	"unicode/utf8"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/image/math/fixed"
)

// LongPressDuration is the default duration of a long press gesture.
// Override this variable to change the detection threshold.
var LongPressDuration time.Duration = 250 * time.Millisecond

// EventType describes a kind of iteraction with rich text.
type EventType uint8

const (
	Hover EventType = iota
	LongPress
	Click
)

// Event describes an interaction with rich text.
type Event struct {
	Type EventType
	// ClickData is only populated if Type == Clicked
	ClickData gesture.ClickEvent
}

// InteractiveSpan holds the persistent state of rich text that can
// be interacted with by the user. It can report clicks, hovers, and
// long-presses on the text.
type InteractiveSpan struct {
	click        gesture.Click
	pressing     bool
	longPressed  bool
	pressStarted time.Time
	contents     string
	metadata     map[string]string
	events       []Event
}

// Layout adds the pointer input op for this interactive span and updates its
// state. It uses the most recent pointer.AreaOp as its input area.
func (i *InteractiveSpan) Layout(gtx layout.Context) layout.Dimensions {
	i.click.Add(gtx.Ops)
	for _, e := range i.click.Events(gtx) {
		switch e.Type {
		case gesture.TypeClick:
			if i.longPressed {
				i.longPressed = false
			} else {
				i.events = append(i.events, Event{Type: Click, ClickData: e})
			}
			i.pressing = false
		case gesture.TypePress:
			i.pressStarted = gtx.Now
			i.pressing = true
		case gesture.TypeCancel:
			i.pressing = false
			i.longPressed = false
		}
	}
	if i.click.Hovered() {
		i.events = append(i.events, Event{Type: Hover})
	}

	if !i.longPressed && i.pressing && gtx.Now.Sub(i.pressStarted) > LongPressDuration {
		i.events = append(i.events, Event{Type: LongPress})
		i.longPressed = true
	}

	if i.pressing && !i.longPressed {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	return layout.Dimensions{}
}

// Events returns click event information for this span.
func (i *InteractiveSpan) Events() []Event {
	out := i.events
	i.events = i.events[:0]
	return out
}

// Content returns the text content of the interactive span as well as the
// metadata associated with it.
func (i *InteractiveSpan) Content() (string, map[string]string) {
	return i.contents, i.metadata
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
	if i != nil {
		i.current = 0
	}
}

// Events returns the first span with unprocessed events and the events that
// need processing for it.
func (i *InteractiveText) Events() (*InteractiveSpan, []Event) {
	for k := range i.Spans {
		span := &i.Spans[k]
		if events := span.Events(); len(events) > 0 {
			return span, events
		}
	}
	return nil, nil
}

// SpanStyle describes the appearance of a span of styled text.
type SpanStyle struct {
	Font        text.Font
	Size        unit.Sp
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
// retrieved if the span is interacted with. If the provided value
// is empty, the key will be deleted from the metadata.
func (ss *SpanStyle) Set(key, value string) {
	if value == "" {
		if ss.metadata != nil {
			delete(ss.metadata, key)
			if len(ss.metadata) == 0 {
				ss.metadata = nil
			}
		}
		return
	}
	if ss.metadata == nil {
		ss.metadata = make(map[string]string)
	}
	ss.metadata[key] = value
}

// Layout renders the span using the provided text shaping.
func (ss SpanStyle) Layout(gtx layout.Context, s text.Shaper, shape spanShape) layout.Dimensions {
	paint.ColorOp{Color: ss.Color}.Add(gtx.Ops)
	defer op.Offset(shape.offset).Push(gtx.Ops).Pop()
	defer clip.Outline{Path: s.Shape(ss.Font, fixed.I(gtx.Sp(ss.Size)), shape.layout)}.Op().Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: shape.size}
}

// DeepCopy returns an identical SpanStyle with its own copy of its metadata.
func (ss SpanStyle) DeepCopy() SpanStyle {
	out := ss
	if len(ss.metadata) > 0 {
		md := make(map[string]string)
		for k, v := range ss.metadata {
			md[k] = v
		}
		out.metadata = md
	}
	return out
}

// TextStyle presents rich text.
type TextStyle struct {
	State  *InteractiveText
	Styles []SpanStyle
	text.Shaper
}

// Text constructs a TextStyle.
func Text(state *InteractiveText, shaper text.Shaper, styles ...SpanStyle) TextStyle {
	return TextStyle{
		State:  state,
		Styles: styles,
		Shaper: shaper,
	}
}

// Layout renders the TextStyle.
func (t TextStyle) Layout(gtx layout.Context) layout.Dimensions {
	spans := make([]SpanStyle, len(t.Styles))
	copy(spans, t.Styles)
	t.State.reset()

	var (
		lineDims       image.Point
		lineAscent     int
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
		lines := t.Shaper.LayoutString(span.Font, fixed.I(gtx.Sp(span.Size)), maxWidth, gtx.Locale, span.Content)

		// grab the first line of the result and compute its dimensions
		firstLine := lines[0]
		spanWidth := firstLine.Width.Ceil()
		spanHeight := (firstLine.Ascent + firstLine.Descent).Ceil()
		spanAscent := firstLine.Ascent.Ceil()

		// forceToNextLine handles the case in which the first segment of the new span does not fit
		// AND there is already content on the current line. If there is no content on the line,
		// we should display the content that doesn't fit anyway, as it won't fit on the next
		// line either.
		forceToNextLine := lineDims.X > 0 && spanWidth > maxWidth

		if !forceToNextLine {
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
			if lineAscent < spanAscent {
				lineAscent = spanAscent
			}

			// update the width of the overall text
			if overallSize.X < lineDims.X {
				overallSize.X = lineDims.X
			}

		}

		// if we are breaking the current span across lines or we are on the
		// last span, lay out all of the spans for the line.
		if len(lines) > 1 || i == len(spans)-1 || forceToNextLine {
			for i, shape := range lineShapes {
				// lay out this span
				span = spans[i+lineStartIndex]
				shape.offset.Y = overallSize.Y + lineAscent
				span.Layout(gtx, t.Shaper, shape)

				if !span.Interactive {
					state = nil
					continue
				}
				// grab an interactive state and lay it out atop the text.
				// If we still have a state, this line is a continuation of
				// the previous span and we should use the same state.
				if state == nil {
					state = t.State.next()
					state.contents = span.Content
					state.metadata = span.metadata
				}
				// set this offset to the upper corner of the text, not the lower
				shape.offset.Y -= lineDims.Y
				offStack := op.Offset(shape.offset).Push(gtx.Ops)
				pr := clip.Rect(image.Rectangle{Max: shape.size}).Push(gtx.Ops)
				state.Layout(gtx)
				pointer.CursorPointer.Add(gtx.Ops)
				pr.Pop()
				offStack.Pop()
				// ensure that we request new state for each interactive text
				// that isn't breaking across a line.
				if i < len(lineShapes)-1 {
					state = nil
				}
			}
			// reset line shaping data and update overall vertical dimensions
			lineShapes = lineShapes[:0]
			overallSize.Y += lineDims.Y
		}

		// if the current span breaks across lines
		if len(lines) > 1 && !forceToNextLine {
			// mark where the next line to be laid out starts
			lineStartIndex = i + 1
			lineDims = image.Point{}
			lineAscent = 0

			// if this span isn't interactive, don't use the same interaction
			// state on the next line.
			if !span.Interactive {
				state = nil
			}

			// ensure the spans slice has room for another span
			spans = append(spans, SpanStyle{})
			// shift existing spans further
			for k := len(spans) - 1; k > i+1; k-- {
				spans[k] = spans[k-1]
			}
			// synthesize and insert a new span
			byteLen := 0
			for i := 0; i < firstLine.Layout.Runes.Count; i++ {
				_, n := utf8.DecodeRuneInString(span.Content[byteLen:])
				byteLen += n
			}
			span.Content = span.Content[byteLen:]
			spans[i+1] = span
		} else if forceToNextLine {
			// mark where the next line to be laid out starts
			lineStartIndex = i
			lineDims = image.Point{}
			lineAscent = 0
			i--
		}
	}

	return layout.Dimensions{Size: gtx.Constraints.Constrain(overallSize)}
}
