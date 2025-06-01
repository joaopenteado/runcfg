package otelcfg

import (
	"context"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
)

type meterProviderConfig struct {
	reader        metric.Reader
	metricOptions []metric.Option
}

type MeterProviderOption = option[meterProviderConfig]

// WithReader sets the reader for the meter provider.
func WithReader(reader metric.Reader) MeterProviderOption {
	return optionFunc[meterProviderConfig](func(cfg *meterProviderConfig) {
		cfg.reader = reader
	})
}

// WithMetricOptions sets the options for the meter provider.
func WithMetricOptions(opts ...metric.Option) MeterProviderOption {
	return optionFunc[meterProviderConfig](func(cfg *meterProviderConfig) {
		cfg.metricOptions = append(cfg.metricOptions, opts...)
	})
}

func SetupMeterProvider(ctx context.Context, opts ...MeterProviderOption) (*metric.MeterProvider, error) {
	cfg := meterProviderConfig{
		reader: nil,
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.reader == nil {
		reader, err := autoexport.NewMetricReader(ctx, autoexport.WithFallbackMetricReader(
			func(ctx context.Context) (metric.Reader, error) {
				return CloudMonitoringMetricReader(ctx)
			},
		))
		if err != nil {
			return nil, err
		}
		cfg.reader = reader
	}

	mpOpts := make([]metric.Option, 1, 1+len(cfg.metricOptions))
	mpOpts[0] = metric.WithReader(cfg.reader)
	mpOpts = append(mpOpts, cfg.metricOptions...)

	mp := metric.NewMeterProvider(mpOpts...)

	otel.SetMeterProvider(mp)

	return mp, nil
}

type cloudMonitoringMetricReaderConfig struct {
	exportOptions []mexporter.Option
	readerOptions []metric.PeriodicReaderOption
}

type CloudMonitoringMetricReaderOption = option[cloudMonitoringMetricReaderConfig]

// WithExporterOptions sets the options for the Cloud Monitoring metric exporter.
func WithExporterOptions(opts ...mexporter.Option) CloudMonitoringMetricReaderOption {
	return optionFunc[cloudMonitoringMetricReaderConfig](func(cfg *cloudMonitoringMetricReaderConfig) {
		cfg.exportOptions = append(cfg.exportOptions, opts...)
	})
}

// WithReaderOptions sets the options for the Cloud Monitoring metric reader.
func WithReaderOptions(opts ...metric.PeriodicReaderOption) CloudMonitoringMetricReaderOption {
	return optionFunc[cloudMonitoringMetricReaderConfig](func(cfg *cloudMonitoringMetricReaderConfig) {
		cfg.readerOptions = append(cfg.readerOptions, opts...)
	})
}

// CloudMonitoringMetricReader creates a new Cloud Monitoring metric reader.
func CloudMonitoringMetricReader(ctx context.Context, opts ...CloudMonitoringMetricReaderOption) (metric.Reader, error) {
	cfg := cloudMonitoringMetricReaderConfig{
		exportOptions: []mexporter.Option{},
	}

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	exporter, err := mexporter.New(cfg.exportOptions...)
	if err != nil {
		return nil, err
	}

	return metric.NewPeriodicReader(exporter, cfg.readerOptions...), nil
}
