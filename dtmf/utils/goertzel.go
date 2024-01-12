package utils

import "math"

type Goertzel32 struct {
	freq []*goertzel
	mag  []float32
	cplx []complex64
}

type goertzel struct {
	coeff    float64
	cos, sin float64
	q1, q2   float64
	q1i, q2i float64
}

func NewGoertzel32(targetFreqs []uint64, sampleRate, blockSize int) *Goertzel32 {
	freq := make([]*goertzel, len(targetFreqs))
	for i, f := range targetFreqs {
		// k is the closest bucket for the frequency
		k := uint64(0.5 + float64(uint64(blockSize)*f)/float64(sampleRate))
		w := 2.0 * math.Pi * float64(k) / float64(blockSize)
		sin := math.Sin(w)
		cos := math.Cos(w)
		freq[i] = &goertzel{
			coeff: 2.0 * cos,
			cos:   cos,
			sin:   sin,
		}
	}
	return &Goertzel32{
		freq: freq,
		mag:  make([]float32, len(targetFreqs)),
		cplx: make([]complex64, len(targetFreqs)),
	}
}

func (g *Goertzel32) Reset() {
	for _, freq := range g.freq {
		freq.q1 = 0.0
		freq.q2 = 0.0
	}
}

func (g *Goertzel32) Feed(samples []float32) {
	for _, samp := range samples {
		for _, freq := range g.freq {
			q0 := freq.coeff*freq.q1 - freq.q2 + float64(samp)
			freq.q2 = freq.q1
			freq.q1 = q0
		}
	}
}

func (g *Goertzel32) Magnitude() []float32 {
	for i, freq := range g.freq {
		g.mag[i] = float32(freq.q1*freq.q1 + freq.q2*freq.q2 - freq.q1*freq.q2*freq.coeff)
	}
	return g.mag
}

// Sliding Goertzel implements a sliding version of the Goertzel filter.
//
// 	x(n)                                               y(n)
// 	──────┬──────(+)──(+)────────────────┬────────(+)─────▶
// 	      ▼       ▲    ▲ ▼               ▼         ▲
// 	    ┌───┐     │    │  ╲            ┌───┐       │
// 	    │z⁻ⁿ│     │    │   ╲           │z⁻ⁿ│       │
// 	    └─┬─┘     │    │    ╲          └─┬─┘       │
// 	      └─▶(x)──┘    │     ╲           │         │
// 	                   │      (x)◀───────●───────▶(x)
// 	                   │       ▲         │         ▲
// 	                   │       │       ┌─▼─┐       │
// 	                   │  2cos(2πk/N)  │z⁻ⁿ│  -e^(-j2πk/N)
// 	                   │               └─┬─┘
// 	                   └──────(x)◀───────┘
// 	                           ▲
// 	                           │
// 	                          -1
