package utils

import (
	"math"
	"strconv"
)

func HammingWindowF32(output []float32) {
	windowF32(output, []float64{0.53836, 1 - 0.53836})
}

func windowF32(output []float32, a []float64) {
	if len(a) < 1 || len(a) > 4 {
		panic("invalid window length " + strconv.Itoa(len(a)))
	}
	nn := float64(len(output) - 1)
	for n := range output {
		fn := float64(n)
		v := a[0]
		if len(a) > 1 {
			v -= a[1] * math.Cos(2*math.Pi*fn/nn)
		}
		if len(a) > 2 {
			v += a[2] * math.Cos(4*math.Pi*fn/nn)
		}
		if len(a) > 3 {
			v -= a[3] * math.Cos(6*math.Pi*fn/nn)
		}
		output[n] = float32(v)
	}
}
