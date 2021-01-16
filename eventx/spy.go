/*
Package events provides types to help manage Gio events and event routing.
*/
package eventx

import (
	"gioui.org/io/event"
	"gioui.org/layout"
)

// Spy wraps an event.Queue and makes a copy of each event that
// is requested from the queue. These copies can be accessed by
// higher-level logic after laying out widgets that consume
// events.
type Spy struct {
	Queue event.Queue

	events []EventGroup
}

var _ event.Queue = &Spy{}

// EventGroup contains a list of events and the tag that they are
// associated with. It can be used as an event.Queue.
type EventGroup struct {
	event.Tag
	Items []event.Event
}

var _ event.Queue = &EventGroup{}

func (e *EventGroup) Events(tag event.Tag) (out []event.Event) {
	if tag != e.Tag {
		return nil
	}
	out, e.Items = e.Items, nil
	return
}

// Enspy returns a new spy and a copy of the layout.Context configured
// to use that spy wrapped around its original queue.
func Enspy(gtx layout.Context) (*Spy, layout.Context) {
	spy := &Spy{Queue: gtx.Queue}
	gtx.Queue = spy
	return spy, gtx

}

// Events returns the events for a given tag from the wrapped Queue.
func (s *Spy) Events(tag event.Tag) []event.Event {
	events := s.Queue.Events(tag)
	s.events = append(s.events, EventGroup{Tag: tag, Items: events})
	return events
}

// AllEvents returns all events that have been requested via the
// Events() method since the last call to AllEvents().
func (s *Spy) AllEvents() (events []EventGroup) {
	events, s.events = s.events, s.events[:0]
	return events
}

// CombinedQueue combines the results of two queues into one.
type CombinedQueue struct {
	A, B event.Queue
}

var _ event.Queue = &CombinedQueue{}

// Combine configures the provided context so that its event queue is
// a CombinedQueue of its original event queue and the provided
// event queue.
func Combine(gtx layout.Context, queue event.Queue) layout.Context {
	gtx.Queue = CombinedQueue{A: gtx.Queue, B: queue}
	return gtx
}

// Events returns the combined results of the two queues.
func (u CombinedQueue) Events(tag event.Tag) []event.Event {
	out := u.A.Events(tag)
	out = append(out, u.B.Events(tag)...)
	return out
}
