package nvml

import "fmt"

// MockNVML is a test implementation of the NVML Interface.
type MockNVML struct {
	Devices       []MockDevice
	InitError     error
	ShutdownError error
}

// MockDevice holds mock data for a single GPU device.
type MockDevice struct {
	Info    DeviceInfo
	Metrics DeviceMetrics
	Error   error
}

var _ Interface = (*MockNVML)(nil)

func (m *MockNVML) Init() error {
	return m.InitError
}

func (m *MockNVML) Shutdown() error {
	return m.ShutdownError
}

func (m *MockNVML) DeviceCount() (int, error) {
	return len(m.Devices), nil
}

func (m *MockNVML) DeviceInfoByIndex(index int) (DeviceInfo, error) {
	if index < 0 || index >= len(m.Devices) {
		return DeviceInfo{}, fmt.Errorf("invalid device index: %d", index)
	}
	if m.Devices[index].Error != nil {
		return DeviceInfo{}, m.Devices[index].Error
	}
	return m.Devices[index].Info, nil
}

func (m *MockNVML) DeviceMetricsByIndex(index int) (DeviceMetrics, error) {
	if index < 0 || index >= len(m.Devices) {
		return DeviceMetrics{}, fmt.Errorf("invalid device index: %d", index)
	}
	if m.Devices[index].Error != nil {
		return DeviceMetrics{}, m.Devices[index].Error
	}
	return m.Devices[index].Metrics, nil
}
