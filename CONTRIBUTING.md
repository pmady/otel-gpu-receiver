# Contributing to GPU Receiver for OpenTelemetry Collector

We welcome contributions! This document outlines the process for contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a feature branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Run tests: `make test`
6. Commit your changes with a descriptive message
7. Push to your fork and submit a Pull Request

## Development Setup

### Prerequisites

- Go 1.23+
- Make
- golangci-lint (for linting)

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

Tests use a mock NVML interface, so no GPU hardware is required for development.

### Linting

```bash
make lint
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports` for formatting
- Write table-driven tests where applicable
- Keep functions focused and small

## Adding New Metrics

To add a new GPU metric:

1. Add the metric config field to `MetricsConfig` in `config.go`
2. Set the default enabled/disabled state in `defaultMetricsConfig()`
3. Add the NVML data field to `DeviceMetrics` in `internal/nvml/nvml.go`
4. Fetch the metric in `internal/nvml/real.go` using the go-nvml bindings
5. Emit the metric in `scraper.go` inside the `scrape()` method
6. Add test coverage in `scraper_test.go`
7. Update the metrics table in `README.md`

## Commit Messages

Use clear, descriptive commit messages:

```
component: brief description of change

Longer explanation if needed. Wrap at 72 characters.
Reference any related issues.
```

Examples:
- `scraper: add PCIe throughput metrics`
- `config: support per-device metric filtering`
- `docs: update metrics table with new ECC counters`

## Reporting Issues

- Use GitHub Issues to report bugs or request features
- Include steps to reproduce for bugs
- Include your GPU model, driver version, and OS for hardware-specific issues

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
