// Package nvml provides an abstraction layer over the NVIDIA Management Library (NVML).
// This allows the GPU receiver to be tested without requiring actual NVIDIA hardware.
package nvml

// DeviceInfo contains the static properties of a GPU device.
type DeviceInfo struct {
	Index int
	UUID  string
	Name  string
}

// DeviceMetrics contains the runtime metrics scraped from a GPU device.
type DeviceMetrics struct {
	GPUUtilization     uint32 // percentage 0-100
	MemoryUsed         uint64 // bytes
	MemoryTotal        uint64 // bytes
	Temperature        uint32 // celsius
	PowerDraw          uint32 // milliwatts
	PowerLimit         uint32 // milliwatts
	ClockSM            uint32 // MHz
	ClockMemory        uint32 // MHz
	EncoderUtilization uint32 // percentage 0-100
	DecoderUtilization uint32 // percentage 0-100
	PCIeThroughputTx   uint32 // KB/s
	PCIeThroughputRx   uint32 // KB/s
	ECCSingleBitErrors uint64
	ECCDoubleBitErrors uint64
}

// Interface abstracts NVML operations for testability.
type Interface interface {
	// Init initializes the NVML library.
	Init() error

	// Shutdown cleans up NVML resources.
	Shutdown() error

	// DeviceCount returns the number of NVIDIA GPU devices in the system.
	DeviceCount() (int, error)

	// DeviceInfoByIndex returns static info about the GPU at the given index.
	DeviceInfoByIndex(index int) (DeviceInfo, error)

	// DeviceMetricsByIndex returns runtime metrics for the GPU at the given index.
	DeviceMetricsByIndex(index int) (DeviceMetrics, error)
}
