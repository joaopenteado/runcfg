package otelcfg

import (
	"github.com/joaopenteado/runcfg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type resourceConfig struct {
	attrs []attribute.KeyValue
}

type ResourceOption = option[resourceConfig]

// WithAttributes sets additional attributes for the resource.
func WithAttributes(attrs ...attribute.KeyValue) ResourceOption {
	return optionFunc[resourceConfig](func(cfg *resourceConfig) {
		cfg.attrs = append(cfg.attrs, attrs...)
	})
}

// NewServiceResource creates a new OpenTelemetry resource for a Cloud Run service.
func NewServiceResource(metadata *runcfg.Metadata, service *runcfg.Service, opts ...ResourceOption) *resource.Resource {
	cfg := &resourceConfig{}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	attrs := append(cfg.attrs,
		// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/f368d047b7c605a7805094537a6922db36eabcdc/detectors/gcp/detector.go#L35
		semconv.CloudProviderGCP,
		semconv.CloudAccountID(metadata.ProjectID),
		semconv.CloudPlatformGCPCloudRun,
		semconv.FaaSName(service.Name),
		semconv.FaaSVersion(service.Revision),
		semconv.FaaSInstance(metadata.InstanceID),
		semconv.CloudRegion(metadata.Region),
	)

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)
}

// NewJobResource creates a new OpenTelemetry resource for a Cloud Run job.
func NewJobResource(metadata *runcfg.Metadata, job *runcfg.Job, opts ...ResourceOption) *resource.Resource {
	cfg := &resourceConfig{}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	attrs := append(cfg.attrs,
		// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/f368d047b7c605a7805094537a6922db36eabcdc/detectors/gcp/detector.go#L35
		semconv.CloudProviderGCP,
		semconv.CloudAccountID(metadata.ProjectID),
		semconv.CloudPlatformGCPCloudRun,
		semconv.FaaSName(job.Name),
		semconv.FaaSInstance(metadata.InstanceID),
		semconv.GCPCloudRunJobExecution(job.Execution),
		semconv.GCPCloudRunJobTaskIndex(int(job.TaskIndex)),
		semconv.CloudRegion(metadata.Region),
	)

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)
}
