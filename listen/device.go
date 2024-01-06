package listen

import (
	pvrecorder "github.com/Picovoice/pvrecorder/binding/go"
)

func GetMicrophons() (map[string]int, error) {
	result := make(map[string]int)
	if devices, err := pvrecorder.GetAvailableDevices(); err != nil {
		return nil, err
	} else {
		for i, device := range devices {
			result[device] = i
		}
	}

	return result, nil
}
