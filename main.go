package main

import (
	"os"
	"syscall"
	"unsafe"

	custom_error "eveready_socket_remote_detector/error"
	"eveready_socket_remote_detector/rtlsdr"

	"eveready_socket_remote_detector/dsp"

	"eveready_socket_remote_detector/eveready"

	"github.com/mjibson/go-dsp/fft"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

func main() {
	json_config, err := os.ReadFile("config.json")
	custom_error.Fatal(err)

	_, err = syscall.LoadDLL(gjson.Get(string(json_config), "libusb_dll_path").String())
	custom_error.Fatal(err)
	rtl_sdr_dll, err := syscall.LoadDLL(gjson.Get(string(json_config), "rtlsdr_dll_path").String())
	custom_error.Fatal(err)

	_, err = rtlsdr.GetRtlSdrDeviceCount(rtl_sdr_dll)
	custom_error.Fatal(err)

	rtl_sdr_index := int(gjson.Get(string(json_config), "rtl_sdr_index").Int())
	_, err = rtlsdr.GetRtlSdrDeviceName(rtl_sdr_dll, rtl_sdr_index)
	custom_error.Fatal(err)

	rtl_sdr_device, err := rtlsdr.GetRtlSdrDevice(rtl_sdr_dll, rtl_sdr_index)
	custom_error.Fatal(err)

	err = rtlsdr.SetRtlSdrCenterFrequency(rtl_sdr_dll, rtl_sdr_device, uint32(gjson.Get(string(json_config), "center_frequency").Int()))
	custom_error.Fatal(err)

	rtl_sdr_samp_rate := uint32(gjson.Get(string(json_config), "sample_rate").Int())
	err = rtlsdr.SetRtlSdrSampRate(rtl_sdr_dll, rtl_sdr_device, uint32(rtl_sdr_samp_rate))
	custom_error.Fatal(err)

	err = rtlsdr.ResetRtlSdrBuffer(rtl_sdr_dll, rtl_sdr_device)
	custom_error.Fatal(err)

	ctx := rtlsdr.RtlSdr_Ctx{Samp_Rate: int(rtl_sdr_samp_rate), Rtl_Sdr_Device: rtl_sdr_device, Rtl_Sdr_Dll: rtl_sdr_dll}
	rtlsdr.ReadRtlAsync(rtl_sdr_dll, rtl_sdr_device, syscall.NewCallbackCDecl(readRtlSdrAsyncCallback), unsafe.Pointer(&ctx), 0, 0)

}

func readRtlSdrAsyncCallback(buf *uint8, buf_len uint32, ctx unsafe.Pointer) int {
	my_ctx := *(*rtlsdr.RtlSdr_Ctx)(ctx)

	for i := 0; i < int(buf_len); i += 2 {
		re_ptr := unsafe.Add(unsafe.Pointer(buf), i)
		im_ptr := unsafe.Add(re_ptr, 1)
		re := *(*uint8)(re_ptr)
		im := *(*uint8)(im_ptr)

		re_bin := rtlsdr.GetBinaryRtlSdrIQ(re)
		im_bin := rtlsdr.GetBinaryRtlSdrIQ(im)

		if funk.ContainsInt([]int{re_bin, im_bin}, 1) {
			signal_sample_count := dsp.MicrosecondsToNumberOfSamples(eveready.Eveready_Signal_Microseconds, my_ctx.Samp_Rate)
			if signal_sample_count*2 > int(buf_len)-(i+1) {
				//not enough samples remaining to construct the signal
				break
			}

			samples := func() []complex128 {
				ret := make([]complex128, 0)
				for j := i; j < i+(signal_sample_count*2)-1; j += 2 {
					re_ptr := unsafe.Add(unsafe.Pointer(buf), j)
					im_ptr := unsafe.Add(re_ptr, 1)
					re := *(*uint8)(re_ptr)
					im := *(*uint8)(im_ptr)

					re_signed := rtlsdr.GetSignedRtlSdrIQ(re)
					im_signed := rtlsdr.GetSignedRtlSdrIQ(im)

					ret = append(ret, complex(re_signed, im_signed))
				}
				return ret
			}()

			fft_arr := fft.FFT(samples)
			fft_arr = dsp.ZeroOutFFTDCOffset(fft_arr)
			highest_magnitude_frequency := dsp.GetHighestMagnitudeFrequency(fft_arr, float64(my_ctx.Samp_Rate))
			samples = fft.IFFT(fft_arr)
			samples = dsp.CenterTimeDomainSamples(samples, highest_magnitude_frequency, float64(my_ctx.Samp_Rate))

			samples, err := dsp.LowPassFilterTimeDomainSamples(samples, float64(eveready.Eveready_Signal_Post_Center_Low_Pass_Freq), my_ctx.Samp_Rate, eveready.Eveready_Signal_Post_Center_Low_Pass_Tap_Count)
			custom_error.Fatal(err)
			sample_magnitudes := dsp.GetComplexMagnitudes(samples)
			magnitude_pulse_start_indexes, magnitude_pulse_end_indexes := dsp.GetMagnitudePulseIndexes(sample_magnitudes)
			isValid := eveready.Demodulate(magnitude_pulse_start_indexes, magnitude_pulse_end_indexes, my_ctx.Samp_Rate)

			if !isValid {
				i = i + (len(samples) * 2)
				continue
			}

			break
		}
	}

	return 0
}