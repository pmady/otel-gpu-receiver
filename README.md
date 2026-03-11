# GPU Receiver for OpenTelemetry Collector

An OpenTelemetry Collector receiver that collects NVIDIA GPU metrics natively using NVML (NVIDIA Management Library) — the same library that powers `nvidia-smi`.

This eliminates the need for running a separate DCGM Exporter sidecar and scraping it via the Prometheus receiver. GPU metrics flow directly into the OTel Collector pipeline with proper resource attributes.

**Status:** Alpha

## Why?

Teams running AI inference (vLLM, Triton, TGI) or training workloads on Kubernetes currently need:
1. NVIDIA DCGM Exporter deployed as a sidecar or DaemonSet
2. Prometheus receiver configured to scrape the DCGM endpoint
3. Manual PromQL/relabeling to get metrics into a usable form

This receiver replaces all three with a single native component.

## Metrics

### Enabled by Default

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| `gpu.utilization` | Gauge | `%` | GPU compute utilization |
| `gpu.memory.used` | Gauge | `By` | GPU memory used in bytes |
| `gpu.memory.total` | Gauge | `By` | GPU total memory in bytes |
| `gpu.temperature` | Gauge | `Cel` | GPU temperature in celsius |
| `gpu.power.draw` | Gauge | `W` | GPU power draw in watts |
| `gpu.power.limit` | Gauge | `W` | GPU power limit in watts |
| `gpu.clock.sm` | Gauge | `MHz` | GPU SM clock speed |
| `gpu.clock.memory` | Gauge | `MHz` | GPU memory clock speed |

### Disabled by Default

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| `gpu.encoder.utilization` | Gauge | `%` | GPU encoder utilization |
| `gpu.decoder.utilization` | Gauge | `%` | GPU decoder utilization |
| `gpu.pcie.throughput.tx` | Gauge | `KBy/s` | PCIe TX throughput |
| `gpu.pcie.throughput.rx` | Gauge | `KBy/s` | PCIe RX throughput |
| `gpu.ecc.errors.single_bit` | Sum | `1` | Correctable ECC errors (cumulative) |
| `gpu.ecc.errors.double_bit` | Sum | `1` | Uncorrectable ECC errors (cumulative) |

## Resource Attributes

Each GPU device emits metrics with the following resource attributes:

| Attribute | Type | Description |
|-----------|------|-------------|
| `gpu.vendor` | string | Always `"nvidia"` |
| `gpu.model` | string | GPU model name (e.g., `NVIDIA A100-SXM4-80GB`) |
| `gpu.uuid` | string | Unique GPU identifier |
| `gpu.index` | int | GPU device index |

## Configuration

```yaml
receivers:
  gpu:
    # How often to scrape GPU metrics (default: 10s)
    collection_interval: 10s
    metrics:
      gpu.utilization:
        enabled: true
      gpu.memory.used:
        enabled: true
      gpu.memory.total:
        enabled: true
      gpu.temperature:
        enabled: true
      gpu.power.draw:
        enabled: true
      gpu.power.limit:
        enabled: true
      gpu.clock.sm:
        enabled: true
      gpu.clock.memory:
        enabled: true
      # Optional metrics (disabled by default)
      gpu.encoder.utilization:
        enabled: false
      gpu.decoder.utilization:
        enabled: false
      gpu.pcie.throughput.tx:
        enabled: false
      gpu.pcie.throughput.rx:
        enabled: false
      gpu.ecc.errors:
        enabled: false

exporters:
  otlp:
    endpoint: "localhost:4317"
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [gpu]
      exporters: [otlp]
```

## Prerequisites

- NVIDIA GPU with driver installed
- NVML library available (comes with NVIDIA driver)
- Linux (NVML is Linux-only for server/datacenter GPUs)

## Building

```bash
make build
```

## Testing

The receiver uses an NVML interface abstraction, so tests run without GPU hardware:

```bash
make test
```

## Building a Custom Collector

Use the [OpenTelemetry Collector Builder](https://opentelemetry.io/docs/collector/custom-collector/) to include this receiver in your distribution:

```yaml
# builder-config.yaml
dist:
  name: custom-otelcol
  description: OTel Collector with GPU receiver
  output_path: ./dist

receivers:
  - gomod: github.com/pmady/otel-gpu-receiver v0.1.0

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.115.0

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.115.0
```

## Architecture

```
┌─────────────────────────────────────────┐
│           OTel Collector                │
│                                         │
│  ┌─────────────┐    ┌───────────────┐   │
│  │ GPU Receiver │───▶│  Pipeline    │   │
│  │  (NVML)     │    │  (batch,etc)  │   │
│  └──────┬──────┘    └───────┬───────┘   │
│         │                   │           │
│         ▼                   ▼           │
│  ┌─────────────┐    ┌───────────────┐   │
│  │ NVIDIA GPU  │    │   Exporter    │  │
│  │ (via NVML)  │    │ (OTLP/Prom)  │    │
│  └─────────────┘    └───────────────┘  │
└─────────────────────────────────────────┘
```

## Related

- [OTel Collector Issue #392](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/392) — Original request for GPU metrics support
- [NVIDIA go-nvml](https://github.com/NVIDIA/go-nvml) — Go bindings for NVML
- [NVIDIA DCGM Exporter](https://github.com/NVIDIA/dcgm-exporter) — Current Prometheus-based approach

## License

Apache License 2.0
