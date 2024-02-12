package rtlsdr

import (
	"errors"
	"log"
	"math"
	"strconv"
	"syscall"
	"unsafe"
)

type RtlSdr_Ctx struct {
	Samp_Rate      int
	Rtl_Sdr_Device unsafe.Pointer
	Rtl_Sdr_Dll    *syscall.DLL
}

func GetRtlSdrDeviceCount(rtl_sdr_dll *syscall.DLL) (uint32, error) {
	rtlGetDeviceCount, err := rtl_sdr_dll.FindProc("rtlsdr_get_device_count")
	if err != nil {
		return 0, err
	}
	ret, _, _ := rtlGetDeviceCount.Call()
	rtlDeviceCount := uint32(ret)
	s := "Rtl SDR device count: " + strconv.FormatInt(int64(ret), 10)
	if rtlDeviceCount == 0 {
		return 0, errors.New(s)
	}
	log.Println(s)
	return rtlDeviceCount, nil
}

func GetRtlSdrDeviceName(rtl_sdr_dll *syscall.DLL, rtl_sdr_index int) (string, error) {
	rtlGetDeviceName, err := rtl_sdr_dll.FindProc("rtlsdr_get_device_name")
	if err != nil {
		return "", err
	}
	rtlsdridx32 := uint32(rtl_sdr_index)
	ret, _, _ := rtlGetDeviceName.Call(uintptr(rtlsdridx32))
	rtlDeviceNameBytes := func() []byte {
		count := 0
		var bytes []byte
		for {
			a := unsafe.Add(unsafe.Pointer(ret), count)
			b := (*byte)(a)
			if *b == 0 {
				break
			}
			bytes = append(bytes, *b)
			count++
		}
		return bytes
	}()
	rtlDeviceName := string(rtlDeviceNameBytes[:])
	s := "Selected rtl device name: " + rtlDeviceName
	if rtlDeviceName == "" {
		return "", errors.New(s)
	}
	log.Println(s)
	return rtlDeviceName, nil
}

func GetRtlSdrDevice(rtl_sdr_dll *syscall.DLL, rtl_sdr_index int) (unsafe.Pointer, error) {
	rtlOpen, err := rtl_sdr_dll.FindProc("rtlsdr_open")
	if err != nil {
		return nil, err
	}

	var i int
	uP := unsafe.Pointer(&i)
	ret, _, _ := rtlOpen.Call(uintptr(uP), uintptr(uint32(rtl_sdr_index)))
	status := int(ret)
	if status != rtl_sdr_index {
		return nil, errors.New("Call to open RTL SDR returned " + strconv.FormatInt(int64(status), 10))
	}

	return *(*unsafe.Pointer)(uP), nil

}

func GetRtlSdrDeviceTunerGains(rtl_sdr_dll *syscall.DLL, rtl_sdr_device unsafe.Pointer) ([]int, error) {
	gains := make([]int, 0)

	// rltGetTunerGains, err := rtl_sdr_dll.FindProc("rtlsdr_get_tuner_gains")
	// if err != nil {
	// 	return gains, err
	// }

	// rltGetTunerGain, err := rtl_sdr_dll.FindProc("rtlsdr_get_tuner_gain")
	// if err != nil {
	// 	return gains, err
	// }
	// tuner_gain, _, _ := rltGetTunerGain.Call(uintptr(rtl_sdr_device), uintptr(unsafe.Pointer(nil)))
	// log.Println(int(tuner_gain))
	// number_of_gains_, _, _ := rltGetTunerGains.Call(uintptr(rtl_sdr_device), uintptr(unsafe.Pointer(nil)))
	// number_of_gains := int(number_of_gains_)
	// if number_of_gains <= 0 {
	// 	return gains, errors.New("Couldn't get number of RTL SDR gains")
	// }

	// gains_ := make([]int32, number_of_gains)
	// rltGetTunerGains.Call(uintptr(rtl_sdr_device), uintptr(unsafe.Pointer(&gains_[0])))

	// log.Println(gains_)
	return gains, nil
}

func SetRtlSdrCenterFrequency(rtl_sdr_dll *syscall.DLL, rtl_sdr_device unsafe.Pointer, center_frequency uint32) error {
	rtlSetCenterFreq, err := rtl_sdr_dll.FindProc("rtlsdr_set_center_freq")
	if err != nil {
		return err
	}

	ret, _, _ := rtlSetCenterFreq.Call(uintptr(rtl_sdr_device), uintptr(center_frequency))
	status := int(ret)
	if status != 0 {
		return errors.New("Couldn't set RTL SDR center frequency")
	}

	return nil
}

func SetRtlSdrSampRate(rtl_sdr_dll *syscall.DLL, rtl_sdr_device unsafe.Pointer, samp_rate uint32) error {

	rtlSetSampRate, err := rtl_sdr_dll.FindProc("rtlsdr_set_sample_rate")
	if err != nil {
		return err
	}

	ret, _, _ := rtlSetSampRate.Call(uintptr(rtl_sdr_device), uintptr(samp_rate))
	status := int(ret)
	if status != 0 {
		return errors.New("Couldn't set RTL SDR samp rate")
	}

	return nil
}

func ResetRtlSdrBuffer(rtl_sdr_dll *syscall.DLL, rtl_sdr_device unsafe.Pointer) error {
	rtlResetBuffer, err := rtl_sdr_dll.FindProc("rtlsdr_reset_buffer")
	if err != nil {
		return err
	}

	ret, _, _ := rtlResetBuffer.Call(uintptr(rtl_sdr_device))
	status := int(ret)
	if status != 0 {
		return errors.New("Couldn't reset RTL SDR buffer")
	}

	return nil
}

func ReadRtlAsync(rtl_sdr_dll *syscall.DLL, rtl_sdr_device unsafe.Pointer, cb uintptr, ctx unsafe.Pointer, buf_num uint32, buf_len uint32) error {
	rtlReadAsync, err := rtl_sdr_dll.FindProc("rtlsdr_read_async")
	if err != nil {
		return err
	}

	ret, _, _ := rtlReadAsync.Call(uintptr(rtl_sdr_device), cb, uintptr(ctx), uintptr(buf_num), uintptr(buf_len))
	status := int(ret)
	if status != 0 {
		return errors.New("Couldn't read RTL SDR async")
	}
	return nil
}

func GetBinaryRtlSdrIQ(ui uint8) int {

	fl := GetSignedRtlSdrIQ(ui)
	fl /= 127
	fl = math.Round(fl)

	if fl < 0 {
		fl = 0
	}

	if math.Signbit(fl) {
		fl *= -1
	}

	return int(fl)
}

func GetSignedRtlSdrIQ(ui uint8) float64 {
	f := float64(ui)
	f -= 128.0
	return f
}
