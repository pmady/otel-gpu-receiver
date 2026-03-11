package gpureceiver // import "github.com/pmady/otel-gpu-receiver"

import (
	"go.opentelemetry.io/collector/component"
)

var (
	typeStr   = component.MustNewType("gpu")
	scopeName = "github.com/pmady/otel-gpu-receiver"
)

const (
	MetricsStability = component.StabilityLevelAlpha
)
