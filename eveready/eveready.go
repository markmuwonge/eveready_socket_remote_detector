package eveready

import (
	"eveready_socket_remote_detector/dsp"
	"log"
	"reflect"
)

var Eveready_Signal_Microseconds int = 42000
var Eveready_Signal_Post_Center_Low_Pass_Freq = 10000
var Eveready_Signal_Post_Center_Low_Pass_Tap_Count = 51
var Eveready_Remote_Signal_Repeat_Count = 8
var eveready_signal_pulse_count = 34

var signals = map[string][]string{
	"a_on":    {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "11", "0", "11", "0", "1", "00", "1"},
	"a_off":   {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "11", "0", "1", "00", "1", "00", "1"},
	"b_on":    {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "1", "00", "1"},
	"b_off":   {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "1"},
	"c_on":    {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1"},
	"c_off":   {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "1", "00", "1", "00", "1"},
	"d_on":    {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "1"},
	"d_off":   {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "1"},
	"all_on":  {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1"},
	"all_off": {"11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "1", "00", "11", "0", "11", "0", "1", "00", "1", "00", "1", "00", "1", "00", "1"},
}

func Demodulate(magnitude_pulse_start_indexes []int, magnitude_pulse_end_indexes []int, sample_rate int) bool {
	ret := false

	if len(magnitude_pulse_start_indexes) == eveready_signal_pulse_count && len(magnitude_pulse_end_indexes) == eveready_signal_pulse_count {
		signal := make([]string, 0)
		for i, v := range magnitude_pulse_start_indexes {
			pulse_width := magnitude_pulse_end_indexes[i] - v
			pulse_microseconds := dsp.NumberOfSamplesToMicroseconds(pulse_width, sample_rate)
			pulse_symbol := getSignalSymbol(pulse_microseconds, true)
			signal = append(signal, pulse_symbol)

			if i == len(magnitude_pulse_start_indexes)-1 {
				break
			}

			drop_width := magnitude_pulse_start_indexes[i+1] - magnitude_pulse_end_indexes[i]
			drop_microseconds := dsp.NumberOfSamplesToMicroseconds(drop_width, sample_rate)
			drop_symbol := getSignalSymbol(drop_microseconds, false)
			signal = append(signal, drop_symbol)
		}
		for i, v := range signals {
			if reflect.DeepEqual(signal, v) {
				log.Println("Signal", i, "detected")
				ret = true
				break
			}
		}
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
