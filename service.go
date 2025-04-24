package runcfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Service contains environment variables and metadata available to Cloud Run
// services. All variables except PORT are available to all containers. The
// PORT variable is only added to the ingress container. For more details see:
// https://cloud.google.com/run/docs/container-contract#services-env-vars
type Service struct {
	// Metadata contains information from the instance metadata server.
	Metadata

	// The port your HTTP server should listen on.
	// By default, requests are sent to port 8080.
	// Read from `PORT environment variable.
	Port uint16

	// The name of the Cloud Run service being run.
	// Read from `K_SERVICE` environment variable.
	ServiceName string

	// The name of the Cloud Run revision being run.
	// Read from `K_REVISION` environment variable.
	ServiceRevision string

	// The name of the Cloud Run configuration that created the revision.
	// Read from `K_CONFIGURATION` environment variable.
	ServiceConfiguration string
}

// LoadService loads configuration for a Cloud Run service, including both
// environment variables and metadata from the metadata server. It returns a
// Service containing the loaded configuration or an error if the loading
// process fails.
//
// The ctx parameter is used for metadata server requests. The opts parameter
// allows specifying options to configure the loading process.
//
// LoadService will return ErrEnvironmentProcess if environment variable
// processing fails, or ErrMetadataFetch if metadata server requests fail.
func LoadService(ctx context.Context, opts ...LoadOption) (*Service, error) {
	cfg := Service{
		ServiceName:          os.Getenv("K_SERVICE"),
		ServiceRevision:      os.Getenv("K_REVISION"),
		ServiceConfiguration: os.Getenv("K_CONFIGURATION"),
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, errors.Join(ErrEnvironmentProcess, fmt.Errorf("invalid PORT value: %w", err))
		}
		if port == 0 {
			return nil, errors.Join(ErrEnvironmentProcess, errors.New("PORT value cannot be 0"))
		}
		cfg.Port = uint16(port)
	}

	loadOpts := defaultOptions()
	for _, opt := range opts {
		opt(&loadOpts)
	}
	if cfg.Port == 0 {
		// If the PORT environment variable is not set, use the default port
		cfg.Port = loadOpts.defaultPort
	}

	metadata, err := LoadMetadata(ctx, loadOpts.requiredMetadata)
	if err != nil {
		return nil, errors.Join(ErrMetadataFetch, err)
	}

	cfg.Metadata = *metadata
	return &cfg, nil
}
