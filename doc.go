// Package gpureceiver implements an OpenTelemetry Collector receiver that
// collects GPU metrics from NVIDIA GPUs using the NVML (NVIDIA Management Library).
//
// It exposes metrics such as GPU utilization, memory usage, temperature,
// power draw, clock speeds, and ECC errors — replacing the need for a
// separate DCGM Exporter sidecar in Kubernetes GPU workloads.
package gpureceiver // import "github.com/pmady/otel-gpu-receiver"
