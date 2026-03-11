package gpureceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/pmady/otel-gpu-receiver/internal/nvml"
)

func TestNewFactory(t *testing.T) {
	mock := &nvml.MockNVML{}
	f := NewFactoryWithNVML(mock)

	assert.Equal(t, typeStr, f.Type())

	cfg := f.CreateDefaultConfig()
	assert.NotNil(t, cfg)
}

func TestCreateMetricsReceiver(t *testing.T) {
	mock := &nvml.MockNVML{
		Devices: []nvml.MockDevice{
			{
				Info:    nvml.DeviceInfo{Index: 0, UUID: "GPU-0", Name: "A100"},
				Metrics: nvml.DeviceMetrics{GPUUtilization: 50},
			},
		},
	}

	f := NewFactoryWithNVML(mock)
	cfg := f.CreateDefaultConfig()
	settings := receivertest.NewNopSettings(typeStr)

	r, err := f.CreateMetrics(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)
	assert.NotNil(t, r)
}
