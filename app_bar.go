package materials

import (
	"image/color"

	"gioui.org/layout"
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

	actions     []AppBarAction
	overflowBtn widget.Clickable
}

// AppBarAction configures an action in the App Bar's action items.
// The state and icon should not be nil.
type AppBarAction struct {
	Name  string
	Icon  *widget.Icon
	State *widget.Clickable
}

func (a AppBarAction) layout(th *material.Theme, gtx layout.Context) layout.Dimensions {
	btn := material.IconButton(th, a.State, a.Icon)
	btn.Size = unit.Dp(24)
	btn.Inset = layout.Inset{
		Left:   unit.Dp(12),
		Right:  unit.Dp(12),
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
	}
	return btn.Layout(gtx)
}

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
				renderActions := a.actions[:visibleActionItems]
				var actions []layout.FlexChild
				for i := range renderActions {
					action := renderActions[i]
					actions = append(actions, layout.Rigid(func(gtx C) D {
						return action.layout(a.Theme, gtx)
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
}
