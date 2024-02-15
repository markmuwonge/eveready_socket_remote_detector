package eveready

import (
	"eveready_socket_remote_detector/dsp"
	"log"
	"reflect"

	"math/rand"

	"github.com/thoas/go-funk"
)

var Eveready_Signal_Microseconds int = 42000
var Eveready_Signal_Post_Center_Low_Pass_Freq = 10000
var Eveready_Signal_Post_Center_Low_Pass_Tap_Count = 51
var Eveready_Remote_Signal_Repeat_Count = 8
var eveready_signal_pulse_count = 34

var signal_preamble = []string{"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00"}
var signal_payload_middle = []string{"1", "00", "1", "00", "1", "00", "11", "0"}
var signal_prelude = []string{"1", "00", "1"}

var signal_payloads = map[string][]string{
	"a_on": {"11", "0", "11", "0", "11", "0", "11", "0", "11", "0", "11", "0", "11", "0", "11", "0"},

	"a_off": {"11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00"},

	"b_on": {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0"},

	"b_off": {"11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00"},

	"c_on": {"11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0"},

	"c_off": {"11", "0", "1", "00", "11", "0", "1", "00", "11", "0", "1", "00", "11", "0", "1", "00"},

	"d_on": {"1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0"},

	"d_off": {"1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00"},

	"all_on": {"1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00"},

	"all_off": {"11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00"},
}

func Demodulate(magnitude_pulse_start_indexes []int, magnitude_pulse_end_indexes []int, sample_rate int) bool {
	ret := false

	signal := make([]string, 0)

	for i, _ := range magnitude_pulse_start_indexes {
		for i2 := i; i2 < len(magnitude_pulse_start_indexes); i2++ {
			pulse_width := magnitude_pulse_end_indexes[i2] - magnitude_pulse_start_indexes[i2]
			pulse_microseconds := dsp.NumberOfSamplesToMicroseconds(pulse_width, sample_rate)
			pulse_symbol := getSignalSymbol(pulse_microseconds, true)
			signal = append(signal, pulse_symbol)

			if i2 == len(magnitude_pulse_start_indexes)-1 {
				break
			}

			drop_width := magnitude_pulse_start_indexes[i2+1] - magnitude_pulse_end_indexes[i2]
			drop_microseconds := dsp.NumberOfSamplesToMicroseconds(drop_width, sample_rate)
			drop_symbol := getSignalSymbol(drop_microseconds, false)
			signal = append(signal, drop_symbol)
		}
		preamble_check := preambleCheck(signal)
		if preamble_check == -1 {
			break
		} else if preamble_check == 0 {
			signal = nil
			continue
		}

		payloadName, result := payloadFirstHalfCheck(signal)
		if result == -1 {
			break
		} else if result == 0 {
			signal = nil
			continue
		}

		payload_middle := payloadMiddleCheck(signal)
		if payload_middle == -1 {
			break
		} else if payload_middle == 0 {
			signal = nil
			continue
		}

		payload_second_half := payloadSecondHalfCheck(signal, payloadName)
		if payload_second_half == -1 {
			break
		} else if payload_second_half == 0 {
			signal = nil
			continue
		}

		payload_prelude := payloadPreludeCheck(signal, payloadName)
		if payload_prelude == -1 {
			break
		} else if payload_prelude == 0 {
			signal = nil
			continue
		}

		log.Println("Signal matched", payloadName)
		ret = true
		break

	}

	return ret
}

func getSignalSymbol(microseconds int, symbolType bool) string {
	symbol := ""
	if microseconds > 500 && microseconds < 1500 {
		if symbolType {
			symbol = "11"
		} else {
			symbol = "00"
		}
	} else if microseconds < 500 {
		if symbolType {
			symbol = "1"
		} else {
			symbol = "0"
		}
	}
	return symbol
}

func preambleCheck(signal []string) int {

	if len(signal_preamble) > len(signal) {
		return -1
	}
	if !reflect.DeepEqual(signal_preamble, signal[:len(signal_preamble)]) {
		return 0
	}

	return 1
}

func getPayloadNames() []string {
	return funk.Keys(signal_payloads).([]string)
}

func payloadFirstHalfCheck(signal []string) (string, int) {

	signal_copy := append([]string{}, signal...)
	signal_copy = signal_copy[len(signal_preamble):]
	signal_payload_names := getPayloadNames()
	random_payload_name := signal_payload_names[rand.Intn(len(signal_payload_names))]

	if len(signal_payloads[random_payload_name])/2 > len(signal_copy) {
		return "", -1
	}

	first_half_signal_payload_match_name := funk.Reduce(signal_payload_names, func(acc string, signal_payload_name string) string {
		signal_payload := signal_payloads[signal_payload_name]

		first_half_payload_length := len(signal_payloads[random_payload_name]) / 2
		if reflect.DeepEqual(signal_payload[:first_half_payload_length], signal_copy[:first_half_payload_length]) {
			return signal_payload_name
		}
		return acc
	}, "").(string)

	if first_half_signal_payload_match_name == "" {
		return first_half_signal_payload_match_name, 0
	}

	return first_half_signal_payload_match_name, 1

}

func payloadMiddleCheck(signal []string) int {

	signal_copy := append([]string{}, signal...)
	signal_payload_names := getPayloadNames()
	random_payload_name := signal_payload_names[rand.Intn(len(signal_payload_names))]

	signal_copy = signal_copy[len(signal_preamble)+len(signal_payloads[random_payload_name])/2:]
	if len(signal_payload_middle) > len(signal_copy) {
		return -1
	}
	if !reflect.DeepEqual(signal_payload_middle, signal_copy[:len(signal_payload_middle)]) {
		return 0
	}

	return 1
}

func payloadSecondHalfCheck(signal []string, signal_payload_name string) int {
	signal_copy := append([]string{}, signal...)

	signal_copy = signal_copy[len(signal_preamble)+len(signal_payloads[signal_payload_name])/2+len(signal_payload_middle):]

	if len(signal_payloads[signal_payload_name])/2 > len(signal_copy) {
		return -1
	}

	payload_second_half_length := len(signal_payloads[signal_payload_name]) / 2
	if !reflect.DeepEqual(signal_payloads[signal_payload_name][payload_second_half_length:], signal_copy[:payload_second_half_length]) {
		return 0
	}
	return 1
}

func payloadPreludeCheck(signal []string, signal_payload_name string) int {

	signal_copy := append([]string{}, signal...)
	signal_copy = signal_copy[len(signal_preamble)+len(signal_payloads[signal_payload_name])+len(signal_payload_middle):]

	if len(signal_prelude) > len(signal_copy) {

		return -1
	}

	if !reflect.DeepEqual(signal_prelude, signal_copy[:len(signal_prelude)]) {
		return 0
	}

	return 1
}
