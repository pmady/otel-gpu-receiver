package gpureceiver // import "github.com/pmady/otel-gpu-receiver"

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

// Config defines the configuration for the GPU metrics receiver.
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	// Metrics controls which GPU metrics are collected.
	Metrics MetricsConfig `mapstructure:"metrics"`
}

// MetricsConfig allows enabling/disabling individual GPU metrics.
type MetricsConfig struct {
	GPUUtilization      MetricConfig `mapstructure:"gpu.utilization"`
	GPUMemoryUsed       MetricConfig `mapstructure:"gpu.memory.used"`
	GPUMemoryTotal      MetricConfig `mapstructure:"gpu.memory.total"`
	GPUTemperature      MetricConfig `mapstructure:"gpu.temperature"`
	GPUPowerDraw        MetricConfig `mapstructure:"gpu.power.draw"`
	GPUPowerLimit       MetricConfig `mapstructure:"gpu.power.limit"`
	GPUClockSM          MetricConfig `mapstructure:"gpu.clock.sm"`
	GPUClockMemory      MetricConfig `mapstructure:"gpu.clock.memory"`
	GPUEncoderUtil      MetricConfig `mapstructure:"gpu.encoder.utilization"`
	GPUDecoderUtil      MetricConfig `mapstructure:"gpu.decoder.utilization"`
	GPUPCIeThroughputTx MetricConfig `mapstructure:"gpu.pcie.throughput.tx"`
	GPUPCIeThroughputRx MetricConfig `mapstructure:"gpu.pcie.throughput.rx"`
	GPUECCErrors        MetricConfig `mapstructure:"gpu.ecc.errors"`
}

// MetricConfig controls whether an individual metric is enabled.
type MetricConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func defaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		GPUUtilization:      MetricConfig{Enabled: true},
		GPUMemoryUsed:       MetricConfig{Enabled: true},
		GPUMemoryTotal:      MetricConfig{Enabled: true},
		GPUTemperature:      MetricConfig{Enabled: true},
		GPUPowerDraw:        MetricConfig{Enabled: true},
		GPUPowerLimit:       MetricConfig{Enabled: true},
		GPUClockSM:          MetricConfig{Enabled: true},
		GPUClockMemory:      MetricConfig{Enabled: true},
		GPUEncoderUtil:      MetricConfig{Enabled: false},
		GPUDecoderUtil:      MetricConfig{Enabled: false},
		GPUPCIeThroughputTx: MetricConfig{Enabled: false},
		GPUPCIeThroughputRx: MetricConfig{Enabled: false},
		GPUECCErrors:        MetricConfig{Enabled: false},
	}
}

func createDefaultConfig() *Config {
	cfg := scraperhelper.NewDefaultControllerConfig()
	cfg.CollectionInterval = 10 * time.Second

	return &Config{
		ControllerConfig: cfg,
		Metrics:          defaultMetricsConfig(),
	}
}

// Validate checks if the receiver configuration is valid.
func (cfg *Config) Validate() error {
	if cfg.CollectionInterval <= 0 {
		return errors.New("collection_interval must be a positive duration")
	}
	return nil
}
