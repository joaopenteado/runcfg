package otelcfg

import (
	"context"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/oauth"
)

type traceProviderConfig struct {
	textMapPropagator  propagation.TextMapPropagator
	exporter           trace.SpanExporter
	tracerProviderOpts []trace.TracerProviderOption
}

type TracerProviderOption = option[traceProviderConfig]

// WithTextMapPropagator sets the text map propagator for the tracer provider.
// By default, the propagator will be configured by autoprop.
func WithTextMapPropagator(propagator propagation.TextMapPropagator) TracerProviderOption {
	return optionFunc[traceProviderConfig](func(cfg *traceProviderConfig) {
		cfg.textMapPropagator = propagator
	})
}

// WithSpanExporter sets the span exporter for the tracer provider.
// By default, the exporter will be configured by autoexport, falling back to
// the CloudTraceOLTPExporter if no environment-specific exporter is detected.
func WithSpanExporter(exporter trace.SpanExporter) TracerProviderOption {
	return optionFunc[traceProviderConfig](func(cfg *traceProviderConfig) {
		cfg.exporter = exporter
	})
}

// WithTracerProviderOptions sets additional options for the tracer provider.
func WithTracerProviderOptions(opts ...trace.TracerProviderOption) TracerProviderOption {
	return optionFunc[traceProviderConfig](func(cfg *traceProviderConfig) {
		cfg.tracerProviderOpts = append(cfg.tracerProviderOpts, opts...)
	})
}

func SetupTracerProvider(ctx context.Context, opts ...TracerProviderOption) (*trace.TracerProvider, error) {
	var cfg traceProviderConfig
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.textMapPropagator == nil {
		// Configure Context Propagation to use the default W3C traceparent
		// format. This can be overriden by the OTEL_PROPAGATORS environment
		// variable. By default, the default propagators is a composite of the
		// tracecontext and baggage propagators.
		cfg.textMapPropagator = autoprop.NewTextMapPropagator()
	}
	otel.SetTextMapPropagator(cfg.textMapPropagator)

	if cfg.exporter == nil {
		// Create a new span exporter using autoexport, which will automatically
		// detect and use the appropriate exporter based on the environment.
		// If no environment-specific exporter is detected, it will fall back to
		// using the CloudTraceOLTPExporter which exports spans to Google Cloud
		// Trace via the OpenTelemetry Protocol (OTLP).
		exporter, err := autoexport.NewSpanExporter(ctx, autoexport.WithFallbackSpanExporter(CloudTraceOLTPExporter))
		if err != nil {
			return nil, err
		}
		cfg.exporter = exporter
		// Shutting down the tracer provider will also shut down the span
		// exporter. No need to explicitly call the shutdown function.
	}

	tpOpts := make([]trace.TracerProviderOption, 0, 1+len(cfg.tracerProviderOpts))
	tpOpts = append(tpOpts, trace.WithBatcher(cfg.exporter))
	tpOpts = append(tpOpts, cfg.tracerProviderOpts...)

	tp := trace.NewTracerProvider(tpOpts...)

	// Set the tracer provider as the global tracer provider.
	otel.SetTracerProvider(tp)

	return tp, nil
}

// CloudTraceOLTPExporter creates a new span exporter for the Cloud Trace
// Telemetry OpenTelemetry Protocol (OTLP) endpoint.
// https://cloud.google.com/trace/docs/migrate-to-otlp-endpoints#telemetry_replace-go
func CloudTraceOLTPExporter(ctx context.Context) (trace.SpanExporter, error) {
	creds, err := oauth.NewApplicationDefault(ctx, "https://www.googleapis.com/auth/trace.append")
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpointURL("https://telemetry.googleapis.com:443/v1/traces"),
		otlptracegrpc.WithDialOption(grpc.WithPerRPCCredentials(creds)),
	)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
