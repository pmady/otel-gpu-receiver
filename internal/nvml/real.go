package nvml

import (
	gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"
)

// RealNVML is the production implementation that calls the actual NVML library.
type RealNVML struct{}

var _ Interface = (*RealNVML)(nil)

func NewRealNVML() *RealNVML {
	return &RealNVML{}
}

func (r *RealNVML) Init() error {
	ret := gonvml.Init()
	if ret != gonvml.SUCCESS {
		return ret
	}
	return nil
}

func (r *RealNVML) Shutdown() error {
	ret := gonvml.Shutdown()
	if ret != gonvml.SUCCESS {
		return ret
	}
	return nil
}

func (r *RealNVML) DeviceCount() (int, error) {
	count, ret := gonvml.DeviceGetCount()
	if ret != gonvml.SUCCESS {
		return 0, ret
	}
	return count, nil
}

func (r *RealNVML) DeviceInfoByIndex(index int) (DeviceInfo, error) {
	device, ret := gonvml.DeviceGetHandleByIndex(index)
	if ret != gonvml.SUCCESS {
		return DeviceInfo{}, ret
	}

	uuid, ret := device.GetUUID()
	if ret != gonvml.SUCCESS {
		return DeviceInfo{}, ret
	}

	name, ret := device.GetName()
	if ret != gonvml.SUCCESS {
		return DeviceInfo{}, ret
	}

	return DeviceInfo{
		Index: index,
		UUID:  uuid,
		Name:  name,
	}, nil
}

func (r *RealNVML) DeviceMetricsByIndex(index int) (DeviceMetrics, error) {
	device, ret := gonvml.DeviceGetHandleByIndex(index)
	if ret != gonvml.SUCCESS {
		return DeviceMetrics{}, ret
	}

	metrics := DeviceMetrics{}

	// GPU utilization
	utilization, ret := device.GetUtilizationRates()
	if ret == gonvml.SUCCESS {
		metrics.GPUUtilization = utilization.Gpu
	}

	// Memory info
	memInfo, ret := device.GetMemoryInfo()
	if ret == gonvml.SUCCESS {
		metrics.MemoryUsed = memInfo.Used
		metrics.MemoryTotal = memInfo.Total
	}

	// Temperature
	temp, ret := device.GetTemperature(gonvml.TEMPERATURE_GPU)
	if ret == gonvml.SUCCESS {
		metrics.Temperature = temp
	}

	// Power
	power, ret := device.GetPowerUsage()
	if ret == gonvml.SUCCESS {
		metrics.PowerDraw = power
	}

	powerLimit, ret := device.GetPowerManagementLimit()
	if ret == gonvml.SUCCESS {
		metrics.PowerLimit = powerLimit
	}

	// Clock speeds
	smClock, ret := device.GetClockInfo(gonvml.CLOCK_SM)
	if ret == gonvml.SUCCESS {
		metrics.ClockSM = smClock
	}

	memClock, ret := device.GetClockInfo(gonvml.CLOCK_MEM)
	if ret == gonvml.SUCCESS {
		metrics.ClockMemory = memClock
	}

	// Encoder/Decoder utilization
	encUtil, _, ret := device.GetEncoderUtilization()
	if ret == gonvml.SUCCESS {
		metrics.EncoderUtilization = encUtil
	}

	decUtil, _, ret := device.GetDecoderUtilization()
	if ret == gonvml.SUCCESS {
		metrics.DecoderUtilization = decUtil
	}

	// PCIe throughput
	txBytes, ret := device.GetPcieThroughput(gonvml.PCIE_UTIL_TX_BYTES)
	if ret == gonvml.SUCCESS {
		metrics.PCIeThroughputTx = txBytes
	}

	rxBytes, ret := device.GetPcieThroughput(gonvml.PCIE_UTIL_RX_BYTES)
	if ret == gonvml.SUCCESS {
		metrics.PCIeThroughputRx = rxBytes
	}

	// ECC errors
	singleBit, ret := device.GetTotalEccErrors(gonvml.MEMORY_ERROR_TYPE_CORRECTED, gonvml.VOLATILE_ECC)
	if ret == gonvml.SUCCESS {
		metrics.ECCSingleBitErrors = singleBit
	}

	doubleBit, ret := device.GetTotalEccErrors(gonvml.MEMORY_ERROR_TYPE_UNCORRECTED, gonvml.VOLATILE_ECC)
	if ret == gonvml.SUCCESS {
		metrics.ECCDoubleBitErrors = doubleBit
	}

	return metrics, nil
}
