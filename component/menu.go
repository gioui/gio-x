package component

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// SurfaceStyle defines the visual aspects of a material design surface
// with (optionally) rounded corners and a drop shadow.
type SurfaceStyle struct {
	*material.Theme
	// The CornerRadius and Elevation fields of the embedded shadow
	// style also define the corner radius and elevation of the card.
	ShadowStyle
}

func Card(th *material.Theme) SurfaceStyle {
	return SurfaceStyle{
		Theme:       th,
		ShadowStyle: Shadow(unit.Dp(4), unit.Dp(4)),
	}
}

func (c SurfaceStyle) Layout(gtx C, w layout.Widget) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			c.ShadowStyle.Layout(gtx)
			surface := clip.UniformRRect(f32.Rectangle{Max: layout.FPt(gtx.Constraints.Min)}, float32(gtx.Px(c.ShadowStyle.CornerRadius)))
			paint.FillShape(gtx.Ops, c.Theme.Bg, surface.Op(gtx.Ops))
			return D{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}

type MenuItemStyle struct {
	State      *widget.Clickable
	HoverColor color.NRGBA

	LabelInset layout.Inset
	Label      material.LabelStyle

	*widget.Icon
	IconSize  unit.Value
	IconInset layout.Inset

	Hint      material.LabelStyle
	HintInset layout.Inset
}

func MenuItem(th *material.Theme, state *widget.Clickable, label string) MenuItemStyle {
	return MenuItemStyle{
		State: state,
		LabelInset: layout.Inset{
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
		},
		IconSize: unit.Dp(24),
		IconInset: layout.Inset{
			Left: unit.Dp(16),
		},
		HintInset: layout.Inset{
			Right: unit.Dp(16),
		},
		Label:      material.Body1(th, label),
		HoverColor: WithAlpha(th.ContrastBg, 0x30),
	}
}

func (m MenuItemStyle) Layout(gtx C) D {
	min := gtx.Constraints.Min.X
	compact := min == 0
	return material.Clickable(gtx, m.State, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				area := image.Rectangle{
					Max: gtx.Constraints.Min,
				}
				if m.State.Hovered() {
					paint.FillShape(gtx.Ops, m.HoverColor, clip.Rect(area).Op())
				}
				return D{Size: area.Max}
			}),
			layout.Stacked(func(gtx C) D {
				gtx.Constraints.Min.X = min
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if m.Icon == nil {
							return D{}
						}
						return m.IconInset.Layout(gtx, func(gtx C) D {
							return m.Icon.Layout(gtx, m.IconSize)
						})
					}),
					layout.Rigid(func(gtx C) D {
						return m.LabelInset.Layout(gtx, func(gtx C) D {
							return m.Label.Layout(gtx)
						})
					}),
					layout.Flexed(1, func(gtx C) D {
						if compact {
							return D{}
						}
						return D{Size: gtx.Constraints.Min}
					}),
					layout.Rigid(func(gtx C) D {
						if empty := (material.LabelStyle{}); m.Hint == empty {
							return D{}
						}
						return m.HintInset.Layout(gtx, func(gtx C) D {
							return m.Hint.Layout(gtx)
						})
					}),
				)
			}),
		)
	})
}

func MenuHintText(th *material.Theme, label string) material.LabelStyle {
	l := material.Body1(th, label)
	l.Color = WithAlpha(l.Color, 0xaa)
	return l
}

// MenuState holds the state of a menu material design component
// across frames.
type MenuState struct {
	OptionList layout.List
	Options    []func(gtx C) D
}

// MenuStyle defines the presentation of a material design menu component.
type MenuStyle struct {
	*MenuState
	*material.Theme
	// Inset applied around the rendered contents of the state's Options field.
	layout.Inset
	SurfaceStyle
}

// Menu constructs a menu with the provided state and a default Surface behind
// it.
func Menu(th *material.Theme, state *MenuState) MenuStyle {
	m := MenuStyle{
		Theme:        th,
		MenuState:    state,
		SurfaceStyle: Card(th),
		Inset: layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
		},
	}
	m.OptionList.Axis = layout.Vertical
	return m
}

// Layout renders the menu.
func (m MenuStyle) Layout(gtx C) D {
	var fakeOps op.Ops
	originalOps := gtx.Ops
	gtx.Ops = &fakeOps
	maxWidth := 0
	for _, w := range m.Options {
		dims := w(gtx)
		if dims.Size.X > maxWidth {
			maxWidth = dims.Size.X
		}
	}
	gtx.Ops = originalOps
	return m.SurfaceStyle.Layout(gtx, func(gtx C) D {
		return m.Inset.Layout(gtx, func(gtx C) D {
			return m.OptionList.Layout(gtx, len(m.Options), func(gtx C, index int) D {
				gtx.Constraints.Min.X = maxWidth
				gtx.Constraints.Max.X = maxWidth
				return m.Options[index](gtx)
			})
		})
	})
}
