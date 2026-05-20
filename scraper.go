package gpureceiver // import "github.com/pmady/otel-gpu-receiver"

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/pmady/otel-gpu-receiver/internal/nvml"
)

type gpuScraper struct {
	logger *zap.Logger
	config *Config
	nvml   nvml.Interface
}

func newGPUScraper(settings component.TelemetrySettings, cfg *Config, nvmlClient nvml.Interface) *gpuScraper {
	return &gpuScraper{
		logger: settings.Logger,
		config: cfg,
		nvml:   nvmlClient,
	}
}

func (s *gpuScraper) start(_ context.Context, _ component.Host) error {
	s.logger.Info("Starting GPU receiver, initializing NVML")
	if err := s.nvml.Init(); err != nil {
		return fmt.Errorf("failed to initialize NVML: %w", err)
	}
	return nil
}

func (s *gpuScraper) shutdown(_ context.Context) error {
	s.logger.Info("Shutting down GPU receiver, closing NVML")
	if err := s.nvml.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown NVML: %w", err)
	}
	return nil
}

func (s *gpuScraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	md := pmetric.NewMetrics()

	deviceCount, err := s.nvml.DeviceCount()
	if err != nil {
		return md, fmt.Errorf("failed to get GPU device count: %w", err)
	}

	if deviceCount == 0 {
		s.logger.Warn("No NVIDIA GPU devices found")
		return md, nil
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	for i := 0; i < deviceCount; i++ {
		info, err := s.nvml.DeviceInfoByIndex(i)
		if err != nil {
			s.logger.Error("Failed to get device info", zap.Int("gpu_index", i), zap.Error(err))
			continue
		}

		metrics, err := s.nvml.DeviceMetricsByIndex(i)
		if err != nil {
			s.logger.Error("Failed to get device metrics", zap.Int("gpu_index", i), zap.Error(err))
			continue
		}

		rm := md.ResourceMetrics().AppendEmpty()
		res := rm.Resource()
		res.Attributes().PutStr("gpu.vendor", "nvidia")
		res.Attributes().PutStr("gpu.model", info.Name)
		res.Attributes().PutStr("gpu.uuid", info.UUID)
		res.Attributes().PutInt("gpu.index", int64(info.Index))

		sm := rm.ScopeMetrics().AppendEmpty()
		sm.Scope().SetName(scopeName)
		sm.Scope().SetVersion("0.1.0")

		mc := s.config.Metrics

		if mc.GPUUtilization.Enabled {
			addGaugeF64(sm, "gpu.utilization", "GPU compute utilization", "%", now, float64(metrics.GPUUtilization))
		}
		if mc.GPUMemoryUsed.Enabled {
			addGaugeI64(sm, "gpu.memory.used", "GPU memory used", "By", now, int64(metrics.MemoryUsed))
		}
		if mc.GPUMemoryTotal.Enabled {
			addGaugeI64(sm, "gpu.memory.total", "GPU total memory", "By", now, int64(metrics.MemoryTotal))
		}
		if mc.GPUTemperature.Enabled {
			addGaugeF64(sm, "gpu.temperature", "GPU temperature", "Cel", now, float64(metrics.Temperature))
		}
		if mc.GPUPowerDraw.Enabled {
			// NVML reports milliwatts, convert to watts
			addGaugeF64(sm, "gpu.power.draw", "GPU power draw", "W", now, float64(metrics.PowerDraw)/1000.0)
		}
		if mc.GPUPowerLimit.Enabled {
			addGaugeF64(sm, "gpu.power.limit", "GPU power limit", "W", now, float64(metrics.PowerLimit)/1000.0)
		}
		if mc.GPUClockSM.Enabled {
			addGaugeI64(sm, "gpu.clock.sm", "GPU SM clock speed", "MHz", now, int64(metrics.ClockSM))
		}
		if mc.GPUClockMemory.Enabled {
			addGaugeI64(sm, "gpu.clock.memory", "GPU memory clock speed", "MHz", now, int64(metrics.ClockMemory))
		}
		if mc.GPUEncoderUtil.Enabled {
			addGaugeF64(sm, "gpu.encoder.utilization", "GPU encoder utilization", "%", now, float64(metrics.EncoderUtilization))
		}
		if mc.GPUDecoderUtil.Enabled {
			addGaugeF64(sm, "gpu.decoder.utilization", "GPU decoder utilization", "%", now, float64(metrics.DecoderUtilization))
		}
		if mc.GPUPCIeThroughputTx.Enabled {
			addGaugeI64(sm, "gpu.pcie.throughput.tx", "GPU PCIe TX throughput", "KBy/s", now, int64(metrics.PCIeThroughputTx))
		}
		if mc.GPUPCIeThroughputRx.Enabled {
			addGaugeI64(sm, "gpu.pcie.throughput.rx", "GPU PCIe RX throughput", "KBy/s", now, int64(metrics.PCIeThroughputRx))
		}
		if mc.GPUECCErrors.Enabled {
			addCumulativeI64(sm, "gpu.ecc.errors.single_bit", "GPU single-bit ECC errors", "1", now, int64(metrics.ECCSingleBitErrors))
			addCumulativeI64(sm, "gpu.ecc.errors.double_bit", "GPU double-bit ECC errors", "1", now, int64(metrics.ECCDoubleBitErrors))
		}
	}

	return md, nil
}

func addGaugeF64(sm pmetric.ScopeMetrics, name, description, unit string, ts pcommon.Timestamp, value float64) {
	m := sm.Metrics().AppendEmpty()
	m.SetName(name)
	m.SetDescription(description)
	m.SetUnit(unit)
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(ts)
	dp.SetDoubleValue(value)
}

func addGaugeI64(sm pmetric.ScopeMetrics, name, description, unit string, ts pcommon.Timestamp, value int64) {
	m := sm.Metrics().AppendEmpty()
	m.SetName(name)
	m.SetDescription(description)
	m.SetUnit(unit)
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(ts)
	dp.SetIntValue(value)
}

func addCumulativeI64(sm pmetric.ScopeMetrics, name, description, unit string, ts pcommon.Timestamp, value int64) {
	m := sm.Metrics().AppendEmpty()
	m.SetName(name)
	m.SetDescription(description)
	m.SetUnit(unit)
	sum := m.SetEmptySum()
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := sum.DataPoints().AppendEmpty()
	dp.SetTimestamp(ts)
	dp.SetIntValue(value)
}
