// Package richtext provides rendering of text containing multiple fonts, styles, and levels of interactivity.
package richtext

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/x/styledtext"
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
	metadata     map[string]interface{}
	events       []Event
}

// Layout adds the pointer input op for this interactive span and updates its
// state. It uses the most recent pointer.AreaOp as its input area.
func (i *InteractiveSpan) Layout(gtx layout.Context) layout.Dimensions {
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()

	pointer.CursorPointer.Add(gtx.Ops)
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
func (i *InteractiveSpan) Content() (string, map[string]interface{}) {
	return i.contents, i.metadata
}

// Get looks up a metadata property on the interactive span.
func (i *InteractiveSpan) Get(key string) interface{} {
	return i.metadata[key]
}

// InteractiveText holds persistent state for a block of text containing
// spans that may be interactive.
type InteractiveText struct {
	Spans []InteractiveSpan
}

// resize makes sure that there are exactly n interactive spans.
func (i *InteractiveText) resize(n int) {
	if n == 0 && i == nil {
		return
	}

	if cap(i.Spans) >= n {
		i.Spans = i.Spans[:n]
	} else {
		i.Spans = make([]InteractiveSpan, n)
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
	Font           font.Font
	Size           unit.Sp
	Color          color.NRGBA
	Content        string
	Interactive    bool
	metadata       map[string]interface{}
	interactiveIdx int
}

// Set configures a metadata key-value pair on the span that can be
// retrieved if the span is interacted with. If the provided value
// is empty, the key will be deleted from the metadata.
func (ss *SpanStyle) Set(key string, value interface{}) {
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
		ss.metadata = make(map[string]interface{})
	}
	ss.metadata[key] = value
}

// DeepCopy returns an identical SpanStyle with its own copy of its metadata.
func (ss SpanStyle) DeepCopy() SpanStyle {
	out := ss
	if len(ss.metadata) > 0 {
		md := make(map[string]interface{})
		for k, v := range ss.metadata {
			md[k] = v
		}
		out.metadata = md
	}
	return out
}

// TextStyle presents rich text.
type TextStyle struct {
	State     *InteractiveText
	Styles    []SpanStyle
	Alignment text.Alignment
	*text.Shaper
}

// Text constructs a TextStyle.
func Text(state *InteractiveText, shaper *text.Shaper, styles ...SpanStyle) TextStyle {
	return TextStyle{
		State:  state,
		Styles: styles,
		Shaper: shaper,
	}
}

// Layout renders the TextStyle.
func (t TextStyle) Layout(gtx layout.Context) layout.Dimensions {
	// OPT(dh): it'd be nice to avoid this allocation
	styles := make([]styledtext.SpanStyle, len(t.Styles))
	numInteractive := 0
	for i := range t.Styles {
		st := &t.Styles[i]
		if st.Interactive {
			st.interactiveIdx = numInteractive
			numInteractive++
		}
		styles[i] = styledtext.SpanStyle{
			Font:    st.Font,
			Size:    st.Size,
			Color:   st.Color,
			Content: st.Content,
		}
	}
	t.State.resize(numInteractive)

	text := styledtext.Text(t.Shaper, styles...)
	text.Alignment = t.Alignment
	return text.Layout(gtx, func(gtx layout.Context, i int, _ layout.Dimensions) {
		span := &t.Styles[i]
		if !span.Interactive {
			return
		}

		state := &t.State.Spans[span.interactiveIdx]
		state.contents = span.Content
		state.metadata = span.metadata
		state.Layout(gtx)
	})
}
