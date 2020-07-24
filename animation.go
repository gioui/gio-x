package materials

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
)

// VisibilityAnimation holds the animation state for animations that transition between a
// "visible" and "invisible" state for a fixed duration of time.
type VisibilityAnimation struct {
	// How long does the animation last
	time.Duration
	State   VisibilityAnimationState
	Started time.Time
}

// Revealed returns the fraction of the animated entity that should be revealed at the current
// time in the animation. This fraction is computed with linear interpolation.
//
// Revealed should be invoked during every frame that v.Animating() returns true.
//
// If the animation reaches its end this frame, Revealed will transition it to a non-animating
// state automatically.
//
// If the animation is in the process of animating, calling Revealed will automatically add
// an InvalidateOp to the provided layout.Context to ensure that the next frame will be generated
// promptly.
func (v *VisibilityAnimation) Revealed(gtx layout.Context) float32 {
	if v.Animating() {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	if v.Duration == time.Duration(0) {
		v.Duration = time.Second
	}
	progress := float32(gtx.Now.Sub(v.Started).Milliseconds()) / float32(v.Milliseconds())
	if progress >= 1 {
		if v.State == Appearing {
			v.State = Visible
		} else if v.State == Disappearing {
			v.State = Invisible
		}
	}
	switch v.State {
	case Visible:
		return 1
	case Invisible:
		return 0
	case Appearing:
		return progress
	case Disappearing:
		return 1 - progress
	}
	return progress
}

// Visible() returns whether any part of the animated entity should be visible during the
// current animation frame.
func (v VisibilityAnimation) Visible() bool {
	return v.State != Invisible
}

// Animating() returns whether the animation is either in the process of appearsing or
// disappearing.
func (v VisibilityAnimation) Animating() bool {
	return v.State == Appearing || v.State == Disappearing
}

// Appear triggers the animation to begin becoming visible at the provided time. It is
// a no-op if the animation is already visible.
func (v *VisibilityAnimation) Appear(now time.Time) {
	if !v.Visible() && !v.Animating() {
		v.State = Appearing
		v.Started = now
	}
}

// Disappear triggers the animation to begin becoming invisible at the provided time.
// It is a no-op if the animation is already invisible.
func (v *VisibilityAnimation) Disappear(now time.Time) {
	if v.Visible() && !v.Animating() {
		v.State = Disappearing
		v.Started = now
	}
}

// VisibilityAnimationState represents possible states that a VisibilityAnimation can
// be in.
type VisibilityAnimationState int

const (
	Visible VisibilityAnimationState = iota
	Disappearing
	Appearing
	Invisible
)
