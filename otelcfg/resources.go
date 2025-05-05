package otelcfg

import (
	"github.com/joaopenteado/runcfg"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// NewServiceResource creates a new OpenTelemetry resource for a Cloud Run service.
func NewServiceResource(metadata *runcfg.Metadata, service *runcfg.Service) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,

		// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/f368d047b7c605a7805094537a6922db36eabcdc/detectors/gcp/detector.go#L35
		semconv.CloudProviderGCP,
		semconv.CloudAccountID(metadata.ProjectID),
		semconv.CloudPlatformGCPCloudRun,
		semconv.FaaSName(service.Name),
		semconv.FaaSVersion(service.Revision),
		semconv.FaaSInstance(metadata.InstanceID),
		semconv.CloudRegion(metadata.Region),
	)
}

// NewJobResource creates a new OpenTelemetry resource for a Cloud Run job.
func NewJobResource(metadata *runcfg.Metadata, job *runcfg.Job) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,

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
}
