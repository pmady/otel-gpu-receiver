package gpureceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap/zaptest"

	"github.com/pmady/otel-gpu-receiver/internal/nvml"
)

func newTestScraper(t *testing.T, mock *nvml.MockNVML, cfg *Config) *gpuScraper {
	if cfg == nil {
		cfg = createDefaultConfig()
	}
	settings := componenttest.NewNopTelemetrySettings()
	settings.Logger = zaptest.NewLogger(t)
	return newGPUScraper(settings, cfg, mock)
}

func TestScrape_SingleGPU(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{
			{
				Info: nvml.DeviceInfo{
					Index: 0,
					UUID:  "GPU-12345678-abcd-efgh-ijkl-123456789abc",
					Name:  "NVIDIA A100-SXM4-80GB",
				},
				Metrics: nvml.DeviceMetrics{
					GPUUtilization: 75,
					MemoryUsed:     40 * 1024 * 1024 * 1024, // 40 GB
					MemoryTotal:    80 * 1024 * 1024 * 1024, // 80 GB
					Temperature:    62,
					PowerDraw:      285000, // 285W in milliwatts
					PowerLimit:     400000, // 400W
					ClockSM:        1410,
					ClockMemory:    1593,
				},
			},
		},
	}

	s := newTestScraper(t, mock, nil)
	require.NoError(t, s.start(context.Background(), nil))
	defer func() { require.NoError(t, s.shutdown(context.Background())) }()

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	// Should have 1 resource (1 GPU)
	assert.Equal(t, 1, md.ResourceMetrics().Len())

	rm := md.ResourceMetrics().At(0)
	res := rm.Resource()

	// Verify resource attributes
	vendor, ok := res.Attributes().Get("gpu.vendor")
	assert.True(t, ok)
	assert.Equal(t, "nvidia", vendor.Str())

	model, ok := res.Attributes().Get("gpu.model")
	assert.True(t, ok)
	assert.Equal(t, "NVIDIA A100-SXM4-80GB", model.Str())

	uuid, ok := res.Attributes().Get("gpu.uuid")
	assert.True(t, ok)
	assert.Equal(t, "GPU-12345678-abcd-efgh-ijkl-123456789abc", uuid.Str())

	gpuIdx, ok := res.Attributes().Get("gpu.index")
	assert.True(t, ok)
	assert.Equal(t, int64(0), gpuIdx.Int())

	// Default config enables 8 metrics
	sm := rm.ScopeMetrics().At(0)
	assert.Equal(t, 8, sm.Metrics().Len())

	// Verify specific metrics
	metricMap := metricsToMap(sm.Metrics())

	assertGaugeF64(t, metricMap, "gpu.utilization", 75.0)
	assertGaugeI64(t, metricMap, "gpu.memory.used", 40*1024*1024*1024)
	assertGaugeI64(t, metricMap, "gpu.memory.total", 80*1024*1024*1024)
	assertGaugeF64(t, metricMap, "gpu.temperature", 62.0)
	assertGaugeF64(t, metricMap, "gpu.power.draw", 285.0) // milliwatts → watts
	assertGaugeF64(t, metricMap, "gpu.power.limit", 400.0)
	assertGaugeI64(t, metricMap, "gpu.clock.sm", 1410)
	assertGaugeI64(t, metricMap, "gpu.clock.memory", 1593)
}

func TestScrape_MultiGPU(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{
			{
				Info:    nvml.DeviceInfo{Index: 0, UUID: "GPU-0", Name: "A100"},
				Metrics: nvml.DeviceMetrics{GPUUtilization: 50, MemoryUsed: 1000, MemoryTotal: 2000},
			},
			{
				Info:    nvml.DeviceInfo{Index: 1, UUID: "GPU-1", Name: "A100"},
				Metrics: nvml.DeviceMetrics{GPUUtilization: 90, MemoryUsed: 1800, MemoryTotal: 2000},
			},
		},
	}

	s := newTestScraper(t, mock, nil)
	require.NoError(t, s.start(context.Background(), nil))
	defer func() { require.NoError(t, s.shutdown(context.Background())) }()

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, md.ResourceMetrics().Len())

	// Verify each GPU has distinct resource attributes
	uuid0, _ := md.ResourceMetrics().At(0).Resource().Attributes().Get("gpu.uuid")
	uuid1, _ := md.ResourceMetrics().At(1).Resource().Attributes().Get("gpu.uuid")
	assert.Equal(t, "GPU-0", uuid0.Str())
	assert.Equal(t, "GPU-1", uuid1.Str())
}

func TestScrape_NoDevices(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{},
	}

	s := newTestScraper(t, mock, nil)
	require.NoError(t, s.start(context.Background(), nil))
	defer func() { require.NoError(t, s.shutdown(context.Background())) }()

	md, err := s.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestScrape_DisabledMetrics(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{
			{
				Info:    nvml.DeviceInfo{Index: 0, UUID: "GPU-0", Name: "A100"},
				Metrics: nvml.DeviceMetrics{GPUUtilization: 80, MemoryUsed: 1000, MemoryTotal: 2000},
			},
		},
	}

	cfg := createDefaultConfig()
	// Disable everything except utilization
	cfg.Metrics.GPUMemoryUsed.Enabled = false
	cfg.Metrics.GPUMemoryTotal.Enabled = false
	cfg.Metrics.GPUTemperature.Enabled = false
	cfg.Metrics.GPUPowerDraw.Enabled = false
	cfg.Metrics.GPUPowerLimit.Enabled = false
	cfg.Metrics.GPUClockSM.Enabled = false
	cfg.Metrics.GPUClockMemory.Enabled = false

	s := newTestScraper(t, mock, cfg)
	require.NoError(t, s.start(context.Background(), nil))
	defer func() { require.NoError(t, s.shutdown(context.Background())) }()

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	sm := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	assert.Equal(t, 1, sm.Metrics().Len())
	assert.Equal(t, "gpu.utilization", sm.Metrics().At(0).Name())
}

func TestScrape_AllMetricsEnabled(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{
			{
				Info: nvml.DeviceInfo{Index: 0, UUID: "GPU-0", Name: "A100"},
				Metrics: nvml.DeviceMetrics{
					GPUUtilization:     80,
					MemoryUsed:         1000,
					MemoryTotal:        2000,
					Temperature:        65,
					PowerDraw:          300000,
					PowerLimit:         400000,
					ClockSM:            1500,
					ClockMemory:        1200,
					EncoderUtilization: 30,
					DecoderUtilization: 20,
					PCIeThroughputTx:   5000,
					PCIeThroughputRx:   3000,
					ECCSingleBitErrors: 2,
					ECCDoubleBitErrors: 0,
				},
			},
		},
	}

	cfg := createDefaultConfig()
	cfg.Metrics.GPUEncoderUtil.Enabled = true
	cfg.Metrics.GPUDecoderUtil.Enabled = true
	cfg.Metrics.GPUPCIeThroughputTx.Enabled = true
	cfg.Metrics.GPUPCIeThroughputRx.Enabled = true
	cfg.Metrics.GPUECCErrors.Enabled = true

	s := newTestScraper(t, mock, cfg)
	require.NoError(t, s.start(context.Background(), nil))
	defer func() { require.NoError(t, s.shutdown(context.Background())) }()

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	sm := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	// 8 default + 4 optional + 2 ECC counters = 14
	assert.Equal(t, 14, sm.Metrics().Len())
}

func TestStart_NVMLInitError(t *testing.T) {
	mock := &nvml.MockNVML{
		InitError: assert.AnError,
	}

	s := newTestScraper(t, mock, nil)
	err := s.start(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize NVML")
}

// metricsToMap converts pmetric.MetricSlice to a map keyed by metric name.
func metricsToMap(ms pmetric.MetricSlice) map[string]pmetric.Metric {
	result := make(map[string]pmetric.Metric, ms.Len())
	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		result[m.Name()] = m
	}
	return result
}

func assertGaugeF64(t *testing.T, metrics map[string]pmetric.Metric, name string, expected float64) {
	t.Helper()
	m, ok := metrics[name]
	require.True(t, ok, "metric %s not found", name)
	assert.Equal(t, pmetric.MetricTypeGauge, m.Type())
	assert.Equal(t, expected, m.Gauge().DataPoints().At(0).DoubleValue())
}

func assertGaugeI64(t *testing.T, metrics map[string]pmetric.Metric, name string, expected int64) {
	t.Helper()
	m, ok := metrics[name]
	require.True(t, ok, "metric %s not found", name)
	assert.Equal(t, pmetric.MetricTypeGauge, m.Type())
	assert.Equal(t, expected, m.Gauge().DataPoints().At(0).IntValue())
}
