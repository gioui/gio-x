package colorpicker

import (
	"gioui.org/layout"
	"image/color"
)

func NewMux(inputs ...ColorInput) *Mux {
	return &Mux{inputs: inputs}
}

type Mux struct {
	inputs []ColorInput
	color  color.NRGBA
}

func (m *Mux) Layout(gtx layout.Context) layout.Dimensions {
	var children []layout.FlexChild
	for _, input := range m.inputs {
		children = append(children, layout.Rigid(input.Layout))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (m *Mux) Color() color.NRGBA {
	return m.color
}

func (m *Mux) SetColor(col color.NRGBA) {
	m.color = col
	for _, input := range m.inputs {
		input.SetColor(col)
	}
}

func (m *Mux) Changed() bool {
	index := -1
	changed := false
	for i, input := range m.inputs {
		if changed {
			input.SetColor(m.color)
		} else if input.Changed() {
			index = i
			changed = true
			m.color = input.Color()
		}
	}
	if index > 1 {
		for _, input := range m.inputs[:index-1] {
			input.SetColor(m.color)
		}
	}
	return changed
}
