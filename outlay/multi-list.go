package outlay

import (
	"gioui.org/layout"
)

// Segment describes a list segment with an optional header.
//
// [Header] will be invoked first, if supplied.
// [Footer] will be invoked last, if supplied.
//
// The segment must specify how many elements it ranges over since
// the [Element] function does not intrinsically describe how many
// elements it's valid for.
type Segment struct {
	Header  layout.Widget
	Footer  layout.Widget
	Element layout.ListElement
	Range   int
}

// Len returns the range accounting for the option [Header] and [Footer].
func (s Segment) Len() int {
	n := s.Range
	if s.Header != nil {
		n += 1
	}
	if s.Footer != nil {
		n += 1
	}
	return n
}

// Layout the list and header. If [Header] is supplied that is drawn for
// index 0. The index is then adjusted and passed along so that [Element]
// always has an index within its [Range]. [Footer] is drawn after the list
// range.
func (s Segment) Layout(gtx layout.Context, ii int) layout.Dimensions {
	if s.Header != nil {
		if ii == 0 {
			return s.Header(gtx)
		}
		ii--
	}
	if s.Footer != nil {
		if ii >= s.Range {
			return s.Footer(gtx)
		}
	}
	return s.Element(gtx, ii)
}

// MultiList provides a layout composed of [Segment]s. Each segment can
// map directly to some data structure, and will receive indexes within
// its [Range] so that it doesn't need to do any index adjustment.
type MultiList []Segment

// Len computes element length of the entire [MultiList].
func (ml MultiList) Len() int {
	n := 0
	for _, s := range ml {
		n += s.Len()
	}
	return n
}

// Layout the using the corresponding list segment.
// The list segment's [Element] function is invoked with relative offsets so that
// it is always working with an index within its [Range].
func (ml MultiList) Layout(gtx layout.Context, ii int) layout.Dimensions {
	if len(ml) == 0 {
		return layout.Dimensions{}
	}
	idx := ii
	for _, s := range ml {
		if idx < s.Len() {
			return s.Layout(gtx, idx)
		}
		idx -= s.Len()
	}
	return layout.Dimensions{}
}
