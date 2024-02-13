package dsp

import (
	"math"
	"slices"
)

func MicrosecondsToNumberOfSamples(microseconds int, samp_rate int) int {

	f := float64(samp_rate) / float64(1000000)

	f *= float64(microseconds)

	return int(math.Ceil(f))
}

func ZeroOutFFTDCOffset(fft_arr []complex128) []complex128 {
	fft_arr[0] = complex(0, 0)
	return fft_arr
}

func GetHighestMagnitudeFrequency(fft_arr []complex128, samp_rate float64) float64 {
	var highest_magnitude_frequency float64 = 0.0
	var highest_magnitude float64 = 0.0

	for i, v := range fft_arr {
		// if i == 0 {
		// 	continue
		// }
		frequency := ((float64(i) * samp_rate) / float64(len(fft_arr)))

		re := real(v)
		im := imag(v)

		hyp := math.Sqrt(math.Pow(float64(re), 2) + math.Pow(float64(im), 2))

		if i == 1 {
			highest_magnitude_frequency = frequency
			highest_magnitude = hyp

			continue
		}

		if hyp > highest_magnitude {
			highest_magnitude_frequency = frequency
			highest_magnitude = hyp

		}
	}
	return highest_magnitude_frequency
}

func GetComplexMagnitudes(complex_numbers []complex128) []float64 {
	hyps := make([]float64, 0)
	for _, complex_number := range complex_numbers {
		re := real(complex_number)
		im := imag(complex_number)

		hyp := math.Sqrt(math.Pow(float64(re), 2) + math.Pow(float64(im), 2))
		hyps = append(hyps, hyp)
	}
	return hyps
}

func GetMagnitudePulseIndexes(magnitudes []float64) ([]int, []int) {
	largest_magnitude := slices.Max(magnitudes)
	smallest_magnitude := slices.Min(magnitudes)

	pulse_start_indexes := make([]int, 0)
	pulse_end_indexes := make([]int, 0)
	symbol := 0
	for i, v := range magnitudes {
		if math.Abs(largest_magnitude-v) < math.Abs(smallest_magnitude-v) {
			if symbol == 0 {
				pulse_start_indexes = append(pulse_start_indexes, i)
				symbol = 1
			}

		} else {
			if symbol == 1 {
				pulse_end_indexes = append(pulse_end_indexes, i)
				symbol = 0
			}
		}
	}
	return pulse_start_indexes, pulse_end_indexes
}

func NumberOfSamplesToMicroseconds(sample_count int, samp_rate int) int {
	number_of_samples_per_microsecond := MicrosecondsToNumberOfSamples(1, samp_rate)

	return sample_count / number_of_samples_per_microsecond
}
