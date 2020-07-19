package materials

import (
	"image"
	"sync"
	"time"

	"gioui.org/f32"
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
	// init ensures that AppBars constructed using struct literal
	// syntax still have their fields initialized before use.
	init sync.Once

	*material.Theme

	NavigationButton widget.Clickable
	NavigationIcon   *widget.Icon
	Title            string

	normalActions actionGroup
	overflowMenu
	overflowBtn    widget.Clickable
	contextualAnim VisibilityAnimation
}

type actionGroup struct {
	actions           []AppBarAction
	actionAnims       []VisibilityAnimation
	overflow          []OverflowAction
	lastOverflowCount int
}

func (a *actionGroup) layout(gtx C, th *material.Theme, overflowBtn *widget.Clickable) D {
	overflowedActions := len(a.actions)
	gtx.Constraints.Min.Y = 0
	widthDp := float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp
	visibleActionItems := int((widthDp / 48) - 1)
	if visibleActionItems < 0 {
		visibleActionItems = 0
	}
	visibleActionItems = min(visibleActionItems, len(a.actions))
	overflowedActions -= visibleActionItems
	var actions []layout.FlexChild
	for i := range a.actions {
		action := a.actions[i]
		anim := &a.actionAnims[i]
		switch anim.State {
		case Visible:
			if i >= visibleActionItems {
				anim.Disappear(gtx.Now)
			}
		case Invisible:
			if i < visibleActionItems {
				anim.Appear(gtx.Now)
			}
		}
		actions = append(actions, layout.Rigid(func(gtx C) D {
			return action.layout(th, anim, gtx)
		}))
	}
	if len(a.overflow)+overflowedActions > 0 {
		actions = append(actions, layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			btn := material.IconButton(th, overflowBtn, moreIcon)
			btn.Size = unit.Dp(24)
			btn.Inset = layout.UniformInset(unit.Dp(6))
			return overflowButtonInset.Layout(gtx, btn.Layout)
		}))
	}
	a.lastOverflowCount = overflowedActions
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, actions...)
}

type overflowMenu struct {
	VisibilityAnimation
	scrim Scrim
	list  layout.List
}

func (o *overflowMenu) layoutOverflow(gtx C, th *material.Theme, actions *actionGroup) D {
	if !o.Visible() {
		return layout.Dimensions{}
	}
	o.scrim.Layout(gtx, &o.VisibilityAnimation)
	defer op.Push(gtx.Ops).Pop()
	width := gtx.Constraints.Max.X / 2
	gtx.Constraints.Min.X = width
	op.Offset(f32.Pt(float32(width), 0)).Add(gtx.Ops)
	var menuMacro op.MacroOp
	menuMacro = op.Record(gtx.Ops)
	dims := layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			gtx.Constraints.Min.X = width
			paintRect(gtx, gtx.Constraints.Min, th.Color.Hint)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx C) D {
			return o.list.Layout(gtx, len(actions.overflow)+actions.lastOverflowCount, func(gtx C, index int) D {
				var action OverflowAction
				if index < actions.lastOverflowCount {
					action = actions.actions[len(actions.actions)-actions.lastOverflowCount+index].OverflowAction
				} else {
					action = actions.overflow[index-actions.lastOverflowCount]
				}
				return material.Clickable(gtx, action.State, func(gtx C) D {
					gtx.Constraints.Min.X = width
					return layout.Inset{
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
						Left:   unit.Dp(8),
					}.Layout(gtx, func(gtx C) D {
						label := material.Label(th, unit.Dp(18), action.Name)
						label.MaxLines = 1
						return label.Layout(gtx)
					})
				})
			})
		}),
	)
	menuOp := menuMacro.Stop()
	progress := o.Revealed(gtx)
	maxWidth := dims.Size.X
	rect := clip.Rect{
		Max: image.Point{
			X: maxWidth,
			Y: int(float32(dims.Size.Y) * progress),
		},
		Min: image.Point{
			X: maxWidth - int(float32(dims.Size.X)*progress),
			Y: 0,
		},
	}
	rect.Add(gtx.Ops)
	menuOp.Add(gtx.Ops)
	return dims
}

// NewAppBar creates and initializes an App Bar.
func NewAppBar(th *material.Theme) *AppBar {
	ab := &AppBar{
		Theme: th,
	}
	ab.initialize()
	return ab
}

func (a *AppBar) initialize() {
	a.init.Do(func() {
		a.overflowMenu.list.Axis = layout.Vertical
		a.overflowMenu.State = Invisible
		a.contextualAnim.State = Invisible
		a.overflowMenu.Duration = overflowAnimationDuration
		a.contextualAnim.Duration = contextualAnimationDuration
		a.overflowMenu.scrim.FinalAlpha = 82
	})
}

// AppBarAction configures an action in the App Bar's action items.
// The state and icon should not be nil.
type AppBarAction struct {
	OverflowAction
	Icon *widget.Icon
}

const (
	actionAnimationDuration     = time.Millisecond * 250
	contextualAnimationDuration = time.Millisecond * 250
	overflowAnimationDuration   = time.Millisecond * 250
)

var actionButtonInset = layout.Inset{
	Top:    unit.Dp(4),
	Bottom: unit.Dp(4),
}

func (a AppBarAction) layout(th *material.Theme, anim *VisibilityAnimation, gtx layout.Context) layout.Dimensions {
	if !anim.Visible() {
		return layout.Dimensions{}
	}
	animating := anim.Animating()
	var macro op.MacroOp
	if animating {
		macro = op.Record(gtx.Ops)
	}
	btn := material.IconButton(th, a.State, a.Icon)
	btn.Size = unit.Dp(24)
	btn.Inset = layout.UniformInset(unit.Dp(12))
	if !animating {
		return btn.Layout(gtx)
	}
	dims := actionButtonInset.Layout(gtx, btn.Layout)
	btnOp := macro.Stop()
	progress := anim.Revealed(gtx)
	dims.Size.X = int(progress * float32(dims.Size.X))
	// ensure this clip transformation stays local to this function
	defer op.Push(gtx.Ops).Pop()

	clip.Rect{
		Max: dims.Size,
	}.Add(gtx.Ops)
	btnOp.Add(gtx.Ops)
	return dims
}

var overflowButtonInset = layout.Inset{
	Top:    unit.Dp(10),
	Bottom: unit.Dp(10),
}

// OverflowAction is an action that is always in the overflow menu.
type OverflowAction struct {
	Name  string
	State *widget.Clickable
}

func (a *AppBar) updateState(gtx layout.Context) {
	if a.overflowBtn.Clicked() && !a.overflowMenu.Visible() {
		a.overflowMenu.Appear(gtx.Now)
	}
	if a.overflowMenu.scrim.Clicked() {
		a.overflowMenu.Disappear(gtx.Now)
	}
}

// Layout renders the app bar. It will span all available horizontal
// space (gtx.Constraints.Max.X), but has a fixed height.
func (a *AppBar) Layout(gtx layout.Context) layout.Dimensions {
	a.initialize()
	a.updateState(gtx)
	originalMaxY := gtx.Constraints.Max.Y
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(56))
	paintRect(gtx, gtx.Constraints.Max, a.Theme.Color.Primary)

	layout.Flex{
		Alignment: layout.Middle,
	}.Layout(gtx,
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
			return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
				title := material.Body1(a.Theme, a.Title)
				title.Color = a.Theme.Color.InvText
				title.TextSize = unit.Dp(18)
				return title.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return layout.E.Layout(gtx, func(gtx C) D {
				return a.normalActions.layout(gtx, a.Theme, &a.overflowBtn)
			})
		}),
	)
	gtx.Constraints.Max.Y = originalMaxY
	a.overflowMenu.layoutOverflow(gtx, a.Theme, &a.normalActions)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// NavigationClicked returns true when the navigation button has been
// clicked in the last frame.
func (a *AppBar) NavigationClicked() bool {
	return a.NavigationButton.Clicked()
}

// SetActions configures the set of actions available in the
// action item area of the app bar. They will be displayed
// in the order provided (from left to right) and will
// collapse into the overflow menu from right to left. The
// provided OverflowActions will always be in the overflow
// menu in the order provided.
func (a *AppBar) SetActions(actions []AppBarAction, overflows []OverflowAction) {
	a.normalActions.actions = actions
	a.normalActions.actionAnims = make([]VisibilityAnimation, len(actions))
	for i := range a.normalActions.actionAnims {
		a.normalActions.actionAnims[i].Duration = actionAnimationDuration
	}
	a.normalActions.overflow = overflows
}

// TriggerContextual causes the AppBar to transform into a contextual
// App Bar with a different set of actions than normal. If the App Bar
// is already in contextual mode, this will update the state of the
// actions. Otherwise, this has no effect.
func (a *AppBar) TriggerContextual(actions []AppBarAction, overflows []OverflowAction) {
}

// StopContextual causes the AppBar to stop showing a contextual menu
// if one is currently being displayed.
func (a *AppBar) StopContextual() {
}

// CloseOverflowMenu requests that the overflow menu be closed if it is
// open.
func (a *AppBar) CloseOverflowMenu(when time.Time) {
	if a.overflowMenu.Visible() {
		a.overflowMenu.Disappear(when)
	}
}
