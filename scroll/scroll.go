package scroll

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// ScrollbarState holds the persistent state for an area that can
// display a scrollbar.
type ScrollbarState struct {
	track, indicator gesture.Click
	drag             gesture.Drag
	ScrollPosition
}

// ScrollTrackStyle configures the presentation of a track for a scroll area.
type ScrollTrackStyle struct {
	// Padding along the major and minor axis of the scrollbar's
	// track. This is used to keep the scrollbar from touching the
	// edges of the content area.
	MajorPadding, MinorPadding unit.Value
}

// ScrollIndicatorStyle configures the presentation of a scroll indicator.
type ScrollIndicatorStyle struct {
	// The smallest that the scroll indicator is allowed to be along
	// the major axis.
	MajorMinLen unit.Value
	// The width of the scroll indicator across the minor axis.
	MinorWidth unit.Value
	// The normal and hovered colors of the scroll indicator.
	Color, HoverColor color.NRGBA
	// The corner radius of the rectangular indicator. 0 will produce
	// square corners. 0.5*MinorWidth will produce perfectly round
	// corners.
	CornerRadius unit.Value
}

// ScrollPosition describes the position of a scrollable viewport atop
// a finite scrollable region.
type ScrollPosition struct {
	// VisibleStart is the start position of the viewport within the
	// scrollabe region represented as a fraction. It should be in the
	// range [0,1]
	VisibleStart float32
	// VisibleEnd is the end position of the viewport within the scrollable
	// region represented as a fraction. It should be in the range [0,1]
	VisibleEnd float32
}

// ScrollbarStyle configures the presentation of a scrollbar.
type ScrollbarStyle struct {
	Axis      layout.Axis
	State     *ScrollbarState
	Track     ScrollTrackStyle
	Indicator ScrollIndicatorStyle
}

// Scrollbar configures the presentation of a scrollbar using the provided
// theme, state, and positional information.
func Scrollbar(th *material.Theme, state *ScrollbarState, pos ScrollPosition) ScrollbarStyle {
	state.ScrollPosition = pos
	return ScrollbarStyle{
		State: state,
		Track: ScrollTrackStyle{
			MajorPadding: unit.Dp(2),
			MinorPadding: unit.Dp(2),
		},
		Indicator: ScrollIndicatorStyle{
			MajorMinLen:  unit.Dp(8),
			MinorWidth:   unit.Dp(6),
			CornerRadius: unit.Dp(3),
			Color:        th.Palette.Fg,
			HoverColor:   th.Palette.ContrastBg,
		},
	}
}

// Layout renders the scrollbar anchored to the appropriate edge of the
// provided context.
func (s ScrollbarStyle) Layout(gtx layout.Context) layout.Dimensions {
	anchoring := func() layout.Direction {
		if s.Axis == layout.Horizontal {
			return layout.S
		}
		return layout.E
	}()
	return anchoring.Layout(gtx, func(gtx C) D {
		// Convert constraints to an axis-independent form.
		gtx.Constraints.Max = s.Axis.Convert(gtx.Constraints.Max)
		gtx.Constraints.Min = s.Axis.Convert(gtx.Constraints.Min)
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		gtx.Constraints.Min.Y = gtx.Px(unit.Add(gtx.Metric, s.Indicator.MinorWidth, s.Track.MinorPadding))

		// Now that we know the dimensions for the scrollbar track, process events
		// that may have modified the indicator position.
		for _, event := range s.State.track.Events(gtx) {
			if event.Type != gesture.TypeClick ||
				event.Modifiers != key.Modifiers(0) ||
				event.NumClicks > 1 {
				continue
			}
			pos := s.Axis.Convert(image.Point{
				X: int(event.Position.X),
				Y: int(event.Position.Y),
			})
			normalizedPos := float32(pos.X) / float32(gtx.Constraints.Max.X)
			delta := normalizedPos - s.State.VisibleStart
			s.State.VisibleStart += delta
			s.State.VisibleEnd += delta
			op.InvalidateOp{}.Add(gtx.Ops)
		}

		return s.layout(gtx)
	})
}

// layout lays out the scroll track and indicator under the assumption
// that the current gtx is already positioned (and rotated) correctly
// for the current scroll axis.
func (s ScrollbarStyle) layout(gtx C) D {
	inset := layout.Inset{
		Top:    s.Track.MajorPadding,
		Bottom: s.Track.MajorPadding,
		Left:   s.Track.MinorPadding,
		Right:  s.Track.MinorPadding,
	}

	return inset.Layout(gtx, func(gtx C) D {
		// Lay out the clickable track underneath the scroll indicator.
		area := s.Axis.Convert(gtx.Constraints.Min)
		pointer.Rect(image.Rectangle{
			Max: area,
		}).Add(gtx.Ops)
		s.State.track.Add(gtx.Ops)

		// Compute the pixel size and position of the scroll indicator within
		// the track.
		trackLen := float32(gtx.Constraints.Min.X)
		viewStart := s.State.VisibleStart * trackLen
		viewEnd := s.State.VisibleEnd * trackLen
		indicatorLen := unit.Max(gtx.Metric, unit.Px(viewEnd-viewStart), s.Indicator.MajorMinLen)
		indicatorDims := s.Axis.Convert(image.Point{
			X: gtx.Px(indicatorLen),
			Y: gtx.Px(s.Indicator.MinorWidth),
		})
		indicatorDimsF := layout.FPt(indicatorDims)
		radius := float32(gtx.Px(s.Indicator.CornerRadius))

		// Lay out the indicator.
		offset := s.Axis.Convert(image.Pt(int(viewStart), 0))
		defer op.Save(gtx.Ops).Load()
		op.Offset(layout.FPt(offset)).Add(gtx.Ops)
		paint.FillShape(gtx.Ops, s.Indicator.Color, clip.RRect{
			Rect: f32.Rectangle{
				Max: indicatorDimsF,
			},
			SW: radius,
			NW: radius,
			NE: radius,
			SE: radius,
		}.Op(gtx.Ops))

		// Add the indicator pointer hit areas.
		pointer.Rect(image.Rectangle{Max: indicatorDims}).Add(gtx.Ops)
		s.State.drag.Add(gtx.Ops)
		s.State.indicator.Add(gtx.Ops)

		return D{Size: area}
	})
}
