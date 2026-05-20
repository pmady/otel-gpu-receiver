package gpureceiver // import "github.com/pmady/otel-gpu-receiver"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/pmady/otel-gpu-receiver/internal/nvml"
)

// NewFactory creates a factory for the GPU receiver.
func NewFactory() receiver.Factory {
	return NewFactoryWithNVML(nvml.NewRealNVML())
}

// NewFactoryWithNVML creates a factory with a custom NVML implementation (for testing).
func NewFactoryWithNVML(nvmlClient nvml.Interface) receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		func() component.Config {
			return createDefaultConfig()
		},
		receiver.WithMetrics(createMetricsReceiverFunc(nvmlClient), MetricsStability),
	)
}

func createMetricsReceiverFunc(nvmlClient nvml.Interface) receiver.CreateMetricsFunc {
	return func(
		ctx context.Context,
		settings receiver.Settings,
		cfg component.Config,
		nextConsumer consumer.Metrics,
	) (receiver.Metrics, error) {
		gpuCfg := cfg.(*Config)
		s := newGPUScraper(settings.TelemetrySettings, gpuCfg, nvmlClient)

		gpuScraper, err := scraper.NewMetrics(
			s.scrape,
			scraper.WithStart(s.start),
			scraper.WithShutdown(s.shutdown),
		)
		if err != nil {
			return nil, err
		}

		return scraperhelper.NewMetricsController(
			&gpuCfg.ControllerConfig,
			settings,
			nextConsumer,
			scraperhelper.AddMetricsScraper(typeStr, gpuScraper),
		)
	}
}
