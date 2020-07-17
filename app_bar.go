package materials

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var moreIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.NavigationMoreVert)
	return icon
}()

// AppBar implements the material design App Bar documented here:
// https://material.io/components/app-bars-top
//
// TODO(whereswaldon): implement support for RTL layouts
type AppBar struct {
	*material.Theme

	NavigationButton widget.Clickable
	NavigationIcon   *widget.Icon
	Title            string

	actions         []AppBarAction
	actionAnimState []appBarAnimation
	overflowBtn     widget.Clickable
}

// AppBarAction configures an action in the App Bar's action items.
// The state and icon should not be nil.
type AppBarAction struct {
	Name  string
	Icon  *widget.Icon
	State *widget.Clickable
}

const actionAnimationDuration = time.Millisecond * 250

func (a AppBarAction) layout(th *material.Theme, anim *appBarAnimation, gtx layout.Context) layout.Dimensions {
	if anim.state == invisible {
		return layout.Dimensions{}
	}
	animating := anim.state == appearing || anim.state == disappearing
	var macro op.MacroOp
	if animating {
		macro = op.Record(gtx.Ops)
	}
	btn := material.IconButton(th, a.State, a.Icon)
	btn.Size = unit.Dp(24)
	btn.Inset = layout.Inset{
		Left:   unit.Dp(12),
		Right:  unit.Dp(12),
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
	}
	if !animating {
		return btn.Layout(gtx)
	}
	dims := btn.Layout(gtx)
	btnOp := macro.Stop()
	progress := float32(gtx.Now.Sub(anim.started).Milliseconds()) / float32(actionAnimationDuration.Milliseconds())
	if anim.state == appearing {
		dims.Size.X = int(progress * float32(dims.Size.X))
		if progress >= 1 {
			anim.state = visible
		}
	} else { //disappearing
		dims.Size.X = int((1 - progress) * float32(dims.Size.X))
		if progress >= 1 {
			anim.state = invisible
		}
	}

	// ensure this clip transformation stays local to this function
	defer op.Push(gtx.Ops).Pop()

	clip.Rect{
		Max: dims.Size,
	}.Add(gtx.Ops)
	btnOp.Add(gtx.Ops)
	return dims
}

// appBarAnimation holds the animation state for a particular app bar action.
// This facilitates actions appearing and disappearing gracefully as the screen
// resizes.
type appBarAnimation struct {
	state   appBarAnimationState
	started time.Time
}

type appBarAnimationState int

const (
	visible appBarAnimationState = iota
	disappearing
	appearing
	invisible
)

// Layout renders the app bar. It will span all available horizontal
// space (gtx.Constraints.Max.X), but has a fixed height.
func (a *AppBar) Layout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(60))
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: 50})
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(59))
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: 75})
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(58))
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: 100})
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(57))
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: 125})
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(56))
	paintRect(gtx, gtx.Constraints.Max, a.Theme.Color.Primary)
	layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if a.NavigationIcon == nil {
				return layout.Dimensions{}
			}
			button := material.IconButton(a.Theme, &a.NavigationButton, a.NavigationIcon)
			button.Size = unit.Dp(24)
			button.Inset = layout.UniformInset(unit.Dp(16))
			return button.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx C) D {
				title := material.Body1(a.Theme, a.Title)
				title.Color = a.Theme.Color.InvText
				title.TextSize = unit.Dp(18)
				return layout.S.Layout(gtx, title.Layout)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				widthDp := float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp
				visibleActionItems := int((widthDp / 48) - 1)
				if visibleActionItems < 0 {
					visibleActionItems = 0
				}
				visibleActionItems = min(visibleActionItems, len(a.actions))
				var actions []layout.FlexChild
				lastVisibleAction := len(a.actions) - visibleActionItems
				for i := range a.actions {
					action := a.actions[i]
					anim := &a.actionAnimState[i]
					switch anim.state {
					case visible:
						if i < lastVisibleAction {
							anim.state = disappearing
							anim.started = gtx.Now
						}
					case invisible:
						if i >= lastVisibleAction {
							anim.state = appearing
							anim.started = gtx.Now
						}
					}
					actions = append(actions, layout.Rigid(func(gtx C) D {
						return action.layout(a.Theme, anim, gtx)
					}))
				}
				actions = append(actions, layout.Rigid(func(gtx C) D {
					btn := material.IconButton(a.Theme, &a.overflowBtn, moreIcon)
					btn.Size = unit.Dp(24)
					btn.Inset = layout.Inset{
						Left:   unit.Dp(6),
						Right:  unit.Dp(6),
						Top:    unit.Dp(16),
						Bottom: unit.Dp(16),
					}
					return btn.Layout(gtx)
				}))
				return layout.Flex{}.Layout(gtx, actions...)
			})
		}),
	)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *AppBar) NavigationClicked() bool {
	return a.NavigationButton.Clicked()
}

// SetActions configures the set of actions available in the
// action item area of the app bar. They will be displayed
// in the order provided (from left to right) and will
// collapse into the overflow menu from right to left.
func (a *AppBar) SetActions(actions []AppBarAction) {
	a.actions = actions
	a.actionAnimState = make([]appBarAnimation, len(actions))
}
