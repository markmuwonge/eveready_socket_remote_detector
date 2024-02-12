package dsp

import (
	"math"
	"math/cmplx"

	custom_error "eveready_socket_remote_detector/error"

	"github.com/mattetti/audio/dsp/filters"
	"github.com/mattetti/audio/dsp/windows"
	"github.com/thoas/go-funk"
)

func CenterTimeDomainSamples(samples []complex128, center_frequency float64, samp_rate float64) []complex128 {
	shifted_complex_numbers := make([]complex128, 0)
	for i, v := range samples {
		var shift_direction int
		if !math.Signbit(center_frequency) {
			shift_direction = -1
		} else {
			shift_direction = 1
		}
		c := complex(0, float64(shift_direction))
		c *= 2 * math.Pi
		c *= complex(center_frequency, 0)
		c /= complex(samp_rate/float64(i), 0)

		shifted_complex_numbers = append(shifted_complex_numbers, v*cmplx.Exp(c))
	}
	return shifted_complex_numbers
}

func LowPassFilterTimeDomainSamples(samples []complex128, cut_off_frequency float64, samp_rate int, taps int) ([]complex128, error) {
	real_numbers := funk.Map(samples, func(c complex128) float64 {
		return real(c)
	}).([]float64)
	imaginary_numbers := funk.Map(samples, func(c complex128) float64 {
		return imag(c)
	}).([]float64)

	sinc := filters.Sinc{CutOffFreq: cut_off_frequency, SamplingFreq: samp_rate, Taps: taps, Window: windows.Hamming}
	fir := filters.FIR{Sinc: &sinc}

	real_numbers, err := fir.LowPass(real_numbers)
	if err != nil {
		custom_error.Warn(err)
		return nil, err
	}

	imaginary_numbers, err = fir.LowPass(imaginary_numbers)
	if err != nil {
		custom_error.Warn(err)
		return nil, err
	}

	filtered_complex_numbers := make([]complex128, 0)

	for i := 0; i < len(samples); i++ {
		re := real_numbers[i]
		im := imaginary_numbers[i]
		filtered_complex_numbers = append(filtered_complex_numbers, complex(re, im))
	}
	return filtered_complex_numbers, nil
}
