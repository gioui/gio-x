package component_test

import (
	"image/color"
	"strconv"
	"testing"

	materials "gioui.org/x/component"
)

type interpolationTest struct {
	start, end, expected color.NRGBA
	progress             float32
}

func (i interpolationTest) Run(t *testing.T) {
	interp := materials.Interpolate(i.start, i.end, i.progress)
	if interp != i.expected {
		t.Fatalf("expected interpolation with progress %f to be %v, got %v", i.progress, i.expected, interp)
	}
}

func TestInterpolate(t *testing.T) {
	zero := color.NRGBA{}
	fives := color.NRGBA{R: 5, G: 5, B: 5, A: 5}
	tens := color.NRGBA{R: 10, G: 10, B: 10, A: 10}
	blue := color.NRGBA{R: 64, G: 80, B: 180, A: 255}
	black := color.NRGBA{A: 255}
	for i, testCase := range []interpolationTest{
		{
			start:    zero,
			end:      tens,
			expected: zero,
			progress: 0,
		},
		{
			start:    zero,
			end:      tens,
			expected: tens,
			progress: 1,
		},
		{
			start:    zero,
			end:      tens,
			expected: fives,
			progress: .5,
		},
		{
			start:    tens,
			end:      zero,
			expected: fives,
			progress: .5,
		},
		{
			start:    blue,
			end:      black,
			expected: color.NRGBA{R: 32, G: 40, B: 90, A: 255},
			progress: .5,
		},
	} {
		t.Run(strconv.Itoa(i), testCase.Run)
	}
}
