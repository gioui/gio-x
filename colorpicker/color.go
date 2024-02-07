/*
Package colorpicker provides simple widgets for selecting an RGBA color
and for choosing one of a set of colors.

The PickerStyle type can be used to render a colorpicker (the state will be
stored in a State). Colorpickers allow choosing specific RGBA values with
sliders or providing an RGB hex code.

The MuxStyle type can be used to render a color multiplexer (the state will
be stored in a MuxState). Color multiplexers provide a choice from among a
set of colors.
*/
package colorpicker

import (
	"encoding/hex"
	"image"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// MuxState holds the state of a color multiplexer. A color multiplexer allows
// choosing from among a set of colors.
type MuxState struct {
	widget.Enum
	Options        map[string]*color.NRGBA
	OrderedOptions []string
}

// NewMuxState creates a MuxState that will provide choices between
// the MuxOptions given as parameters.
func NewMuxState(options ...MuxOption) MuxState {
	keys := make([]string, 0, len(options))
	mapped := make(map[string]*color.NRGBA)
	for _, opt := range options {
		keys = append(keys, opt.Label)
		mapped[opt.Label] = opt.Value
	}
	state := MuxState{
		Options:        mapped,
		OrderedOptions: keys,
	}
	if len(keys) > 0 {
		state.Enum.Value = keys[0]
	}
	return state
}

// MuxOption is one choice for the value of a color multiplexer.
type MuxOption struct {
	Label string
	Value *color.NRGBA
}

// Color returns the currently-selected color.
func (m MuxState) Color() *color.NRGBA {
	return m.Options[m.Enum.Value]
}

// MuxStyle renders a MuxState as a material design themed widget.
type MuxStyle struct {
	*MuxState
	Theme *material.Theme
	Label string
}

// Mux creates a MuxStyle from a theme and a state.
func Mux(theme *material.Theme, state *MuxState, label string) MuxStyle {
	return MuxStyle{
		Theme:    theme,
		MuxState: state,
		Label:    label,
	}
}

// Layout renders the MuxStyle into the provided context.
func (m MuxStyle) Layout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	var children []layout.FlexChild
	inset := layout.UniformInset(unit.Dp(2))
	children = append(children, layout.Rigid(func(gtx C) D {
		return inset.Layout(gtx, func(gtx C) D {
			return material.Body1(m.Theme, m.Label).Layout(gtx)
		})
	}))
	for i := range m.OrderedOptions {
		opt := m.OrderedOptions[i]
		children = append(children, layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				return m.layoutOption(gtx, opt)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (m MuxStyle) layoutOption(gtx C, option string) D {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.RadioButton(m.Theme, &m.Enum, option, option).Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
				color := m.Options[option]
				if color == nil {
					return D{}
				}
				return borderedSquare(gtx, *color)
			})
		}),
	)
}

func borderedSquare(gtx C, c color.NRGBA) D {
	dims := square(gtx, unit.Dp(20), color.NRGBA{A: 255})

	off := gtx.Dp(unit.Dp(1))
	defer op.Offset(image.Pt(off, off)).Push(gtx.Ops).Pop()
	square(gtx, unit.Dp(18), c)
	return dims
}

func square(gtx C, sizeDp unit.Dp, color color.NRGBA) D {
	return rect(gtx, sizeDp, sizeDp, color)
}

func rect(gtx C, width, height unit.Dp, color color.NRGBA) D {
	w, h := gtx.Dp(width), gtx.Dp(height)
	return rectAbs(gtx, w, h, color)
}

func rectAbs(gtx C, w, h int, color color.NRGBA) D {
	size := image.Point{X: w, Y: h}
	bounds := image.Rectangle{Max: size}
	paint.FillShape(gtx.Ops, color, clip.Rect(bounds).Op())
	return D{Size: image.Pt(w, h)}
}

// State is the state of a colorpicker.
type State struct {
	R, G, B, A widget.Float
	widget.Editor

	changed bool
}

// SetColor changes the color represented by the colorpicker.
func (s *State) SetColor(c color.NRGBA) {
	s.R.Value = float32(c.R) / 255.0
	s.G.Value = float32(c.G) / 255.0
	s.B.Value = float32(c.B) / 255.0
	s.A.Value = float32(c.A) / 255.0
	s.updateEditor()
}

// Color returns the currently selected color.
func (s State) Color() color.NRGBA {
	return color.NRGBA{
		R: s.Red(),
		G: s.Green(),
		B: s.Blue(),
		A: s.Alpha(),
	}
}

// Red returns the red value of the currently selected color.
func (s State) Red() uint8 {
	return uint8(s.R.Value * 255)
}

// Green returns the green value of the currently selected color.
func (s State) Green() uint8 {
	return uint8(s.G.Value * 255)
}

// Blue returns the blue value of the currently selected color.
func (s State) Blue() uint8 {
	return uint8(s.B.Value * 255)
}

// Alpha returns the alpha value of the currently selected color.
func (s State) Alpha() uint8 {
	return uint8(s.A.Value * 255)
}

// Update handles all state updates from the underlying widgets.
func (s *State) Update(gtx layout.Context) bool {
	changed := false
	if s.R.Update(gtx) || s.G.Update(gtx) || s.B.Update(gtx) || s.A.Update(gtx) {
		s.updateEditor()
		changed = true
	}
	for {
		_, ok := s.Editor.Update(gtx)
		if !ok {
			break
		}
		out, err := hex.DecodeString(s.Editor.Text())
		if err == nil && len(out) == 3 {
			s.R.Value = (float32(out[0]) / 255.0)
			s.G.Value = (float32(out[1]) / 255.0)
			s.B.Value = (float32(out[2]) / 255.0)
			changed = true
		}
	}
	return changed
}

func (s *State) updateEditor() {
	s.Editor.SetText(hex.EncodeToString([]byte{s.Red(), s.Green(), s.Blue()}))
}

// PickerStyle renders a color picker using material widgets.
type PickerStyle struct {
	*State
	*material.Theme
	Label string
	// MonospaceFace selects the typeface to use for monospace text fields.
	// The zero value will use the generic family "monospace".
	MonospaceFace font.Typeface
}

type (
	C = layout.Context
	D = layout.Dimensions
)

// Picker creates a pickerstyle from a theme and a state.
func Picker(th *material.Theme, state *State, label string) PickerStyle {
	return PickerStyle{
		Theme: th,
		State: state,
		Label: label,
	}
}

// Layout renders the PickerStyle into the provided context.
func (p PickerStyle) Layout(gtx layout.Context) layout.Dimensions {
	p.State.Update(gtx)

	// lay out the label and editor to compute their width
	leftSide := op.Record(gtx.Ops)
	leftSideDims := p.layoutLeftPane(gtx)
	layoutLeft := leftSide.Stop()

	// lay out the sliders in the remaining horizontal space
	rgtx := gtx
	rgtx.Constraints.Max.X -= leftSideDims.Size.X
	rightSide := op.Record(gtx.Ops)
	rightSideDims := p.layoutSliders(rgtx)
	layoutRight := rightSide.Stop()

	// compute the space beneath the editor that will not extend
	// past the sliders vertically
	margin := gtx.Dp(unit.Dp(4))
	sampleWidth, sampleHeight := leftSideDims.Size.X, rightSideDims.Size.Y-leftSideDims.Size.Y

	// lay everything out for real, starting with the editor/label
	layoutLeft.Add(gtx.Ops)

	// offset downwards and lay out the color sample
	var stack op.TransformStack
	stack = op.Offset(image.Pt(margin, leftSideDims.Size.Y)).Push(gtx.Ops)
	rectAbs(gtx, sampleWidth-(2*margin), sampleHeight-(2*margin), p.State.Color())
	stack.Pop()

	// offset to the right to lay out the sliders
	defer op.Offset(image.Pt(leftSideDims.Size.X, 0)).Push(gtx.Ops).Pop()
	layoutRight.Add(gtx.Ops)

	return layout.Dimensions{
		Size: image.Point{
			X: gtx.Constraints.Max.X,
			Y: rightSideDims.Size.Y,
		},
	}
}

func (p PickerStyle) layoutLeftPane(gtx C) D {
	monospaceFace := p.MonospaceFace
	if len(p.MonospaceFace) == 0 {
		monospaceFace = "monospace"
	}
	gtx.Constraints.Min.X = 0
	inset := layout.UniformInset(unit.Dp(4))
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				return material.Body1(p.Theme, p.Label).Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return rectAbs(gtx, gtx.Constraints.Min.X, gtx.Constraints.Min.Y, color.NRGBA{R: 230, G: 230, B: 230, A: 255})
					}),
					layout.Stacked(func(gtx C) D {
						return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
							return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									label := material.Body1(p.Theme, "#")
									label.Font.Typeface = monospaceFace
									return label.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									editor := material.Editor(p.Theme, &p.Editor, "rrggbb")
									editor.Font.Typeface = monospaceFace
									return editor.Layout(gtx)
								}),
							)
						})
					}),
				)
			})
		}),
	)
	return dims
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p PickerStyle) layoutSliders(gtx C) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.R, "R:", valueString(p.Red()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.G, "G:", valueString(p.Green()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.B, "B:", valueString(p.Blue()))
		}),
		layout.Rigid(func(gtx C) D {
			return p.layoutSlider(gtx, &p.A, "A:", valueString(p.Alpha()))
		}),
	)
}

func valueString(in uint8) string {
	s := strconv.Itoa(int(in))
	delta := 3 - len(s)
	if delta > 0 {
		s = strings.Repeat(" ", delta) + s
	}
	return s
}

func (p PickerStyle) layoutSlider(gtx C, value *widget.Float, label, valueStr string) D {
	monospaceFace := p.MonospaceFace
	if len(p.MonospaceFace) == 0 {
		monospaceFace = "monospace"
	}
	inset := layout.UniformInset(unit.Dp(2))
	layoutDims := layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Body1(p.Theme, label)
				label.Font.Typeface = monospaceFace
				return label.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			sliderDims := inset.Layout(gtx, material.Slider(p.Theme, value).Layout)
			return sliderDims
		}),
		layout.Rigid(func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Body1(p.Theme, valueStr)
				label.Font.Typeface = monospaceFace
				return label.Layout(gtx)
			})
		}),
	)
	return layoutDims
}
