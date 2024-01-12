package dtmf

import (
	"bytes"
	"dtmf-decoder/dtmf/utils"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Keypad = []rune{
		'1', '2', '3', 'A',
		'4', '5', '6', 'B',
		'7', '8', '9', 'C',
		'*', '0', '#', 'D',
	}
	StdLowFreq  = []uint64{697, 770, 852, 941}
	StdHighFreq = []uint64{1209, 1336, 1477, 1633}
)

type DTMF struct {
	lowFreq   *utils.Goertzel32
	highFreq  *utils.Goertzel32
	nHigh     int
	blockSize int
	w         []float32
}

func New(lowFreq, highFreq []uint64, sampleRate, blockSize int, windowFunc func([]float32)) *DTMF {
	w := make([]float32, blockSize)
	if windowFunc != nil {
		windowFunc(w)
	} else {
		utils.HammingWindowF32(w)
	}
	return &DTMF{
		lowFreq:   utils.NewGoertzel32(lowFreq, sampleRate, blockSize),
		highFreq:  utils.NewGoertzel32(highFreq, sampleRate, blockSize),
		nHigh:     len(highFreq),
		blockSize: blockSize,
		w:         w,
	}
}

func NewStandard(sampleRate, blockSize int) *DTMF {
	return New(StdLowFreq, StdHighFreq, sampleRate, blockSize, utils.HammingWindowF32)
}

// Feed Return key number (lowFreqIndex * numHighFreq + highFreqIndex) and minimum magnitude
func (d *DTMF) Feed(samples []float32) (int, float32) {
	if len(samples) > d.blockSize {
		samples = samples[:d.blockSize]
	}
	for i, s := range samples {
		samples[i] = s * d.w[i]
	}
	d.lowFreq.Reset()
	d.highFreq.Reset()
	d.lowFreq.Feed(samples)
	d.highFreq.Feed(samples)

	row, thresh1 := max(d.lowFreq.Magnitude())
	col, thresh2 := max(d.highFreq.Magnitude())
	if thresh2 < thresh1 {
		thresh1 = thresh2
	}
	return row*d.nHigh + col, thresh1
}

func max(val []float32) (int, float32) {
	lrg := float32(0.0)
	idx := 0
	for i, f := range val {
		if f > lrg {
			lrg = f
			idx = i
		}
	}
	return idx, lrg
}

func DecodeDTMFFromBytes(audioBytes []byte, rate float64, wiggleRoom int) (string, error) {
	if len(audioBytes) == 0 {
		return "", errors.New("audio in the dtmf structure contains no bytes")
	}

	var dtmfOutput string
	sampleRate := int(rate)
	blockSize := 205 * sampleRate / 8000
	window := blockSize / 4
	dt := NewStandard(sampleRate, blockSize)
	lastKey := -1
	keyCount := 0
	samples := make([]float32, blockSize)

	rd := bytes.NewReader(audioBytes)

	buf := make([]byte, window*2)

	for {
		_, err := rd.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		copy(samples, samples[window:])

		si := len(samples) - window
		for i := 0; i < len(buf); i += 2 {
			s := float32(int16(buf[i])|(int16(buf[i+1])<<8)) / 32768.0
			samples[si] = s
			si++
		}

		if k, t := dt.Feed(samples); k == lastKey && t > 0.0 {
			keyCount++
			if keyCount == wiggleRoom {
				dtmfOutput += string(Keypad[k])
				fmt.Println(string(Keypad[k]))
			}
		} else {
			lastKey = k
			keyCount = 0
		}
	}

	return dtmfOutput, nil
}

func DecodeDTMFFromFile(filepath string, wiggleRoom int) (string, error) {
	audioBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "N/A", err
	}

	//formatTag, numChannels, sampleRate, bitsPerSample, dataStart, err := ParseAudioHeader(audioBytes)
	//fmt.Println(formatTag, numChannels, sampleRate, bitsPerSample, dataStart, err)

	decodedValue, err := DecodeDTMFFromBytes(audioBytes, float64(44100), wiggleRoom)
	if err != nil {
		return "N/A", err
	}
	return decodedValue, nil
}

func ParseAudioHeader(wavhdr []byte) (formatTag uint16, numChannels uint16, sampleRate uint32, bitsPerSample uint16, dataStart int, err error) {
	// WAV header format: https://en.wikipedia.org/wiki/WAV
	if string(wavhdr[0:4]) != "RIFF" || string(wavhdr[8:12]) != "WAVE" {
		err = fmt.Errorf("Invalid WAV header")
		return
	}

	formatTag = binary.LittleEndian.Uint16(wavhdr[20:22])
	numChannels = binary.LittleEndian.Uint16(wavhdr[22:24])
	sampleRate = binary.LittleEndian.Uint32(wavhdr[24:28])
	bitsPerSample = binary.LittleEndian.Uint16(wavhdr[34:36])

	// Find the position of the data chunk
	dataStart = 36
	for string(wavhdr[dataStart:dataStart+4]) != "data" {
		dataSize := int(binary.LittleEndian.Uint32(wavhdr[dataStart+4 : dataStart+8]))
		dataStart += 8 + dataSize
	}

	return
}
