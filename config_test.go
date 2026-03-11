package gpureceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

func TestDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()

	assert.Equal(t, 10*time.Second, cfg.CollectionInterval)

	// Default enabled metrics
	assert.True(t, cfg.Metrics.GPUUtilization.Enabled)
	assert.True(t, cfg.Metrics.GPUMemoryUsed.Enabled)
	assert.True(t, cfg.Metrics.GPUMemoryTotal.Enabled)
	assert.True(t, cfg.Metrics.GPUTemperature.Enabled)
	assert.True(t, cfg.Metrics.GPUPowerDraw.Enabled)
	assert.True(t, cfg.Metrics.GPUPowerLimit.Enabled)
	assert.True(t, cfg.Metrics.GPUClockSM.Enabled)
	assert.True(t, cfg.Metrics.GPUClockMemory.Enabled)

	// Default disabled metrics
	assert.False(t, cfg.Metrics.GPUEncoderUtil.Enabled)
	assert.False(t, cfg.Metrics.GPUDecoderUtil.Enabled)
	assert.False(t, cfg.Metrics.GPUPCIeThroughputTx.Enabled)
	assert.False(t, cfg.Metrics.GPUPCIeThroughputRx.Enabled)
	assert.False(t, cfg.Metrics.GPUECCErrors.Enabled)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			cfg:     createDefaultConfig(),
			wantErr: false,
		},
		{
			name: "valid custom interval",
			cfg: &Config{
				ControllerConfig: func() scraperhelper.ControllerConfig {
					c := scraperhelper.NewDefaultControllerConfig()
					c.CollectionInterval = 30 * time.Second
					return c
				}(),
				Metrics: defaultMetricsConfig(),
			},
			wantErr: false,
		},
		{
			name: "invalid zero interval",
			cfg: &Config{
				ControllerConfig: func() scraperhelper.ControllerConfig {
					c := scraperhelper.NewDefaultControllerConfig()
					c.CollectionInterval = 0
					return c
				}(),
				Metrics: defaultMetricsConfig(),
			},
			wantErr: true,
		},
		{
			name: "invalid negative interval",
			cfg: &Config{
				ControllerConfig: func() scraperhelper.ControllerConfig {
					c := scraperhelper.NewDefaultControllerConfig()
					c.CollectionInterval = -5 * time.Second
					return c
				}(),
				Metrics: defaultMetricsConfig(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
