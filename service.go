package runcfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Service contains environment variables available to Cloud Run services.
// All variables except PORT are available to all containers. The PORT variable
// is only added to the ingress container. For more details see the [container
// runtime contract for Services].
//
// [container runtime contract for Services]: https://cloud.google.com/run/docs/container-contract#services-env-vars
type Service struct {
	// The port your HTTP server should listen on.
	// By default, requests are sent to port 8080.
	// Read from `PORT environment variable.
	Port uint16

	// The name of the Cloud Run service being run.
	// Read from `K_SERVICE` environment variable.
	Name string

	// The name of the Cloud Run revision being run.
	// Read from `K_REVISION` environment variable.
	Revision string

	// The name of the Cloud Run configuration that created the revision.
	// Read from `K_CONFIGURATION` environment variable.
	Configuration string
}

func defaultService() *Service {
	return &Service{
		Port: 8080,
	}
}

type ServiceLoadOption func(*Service)

// WithDefaultPort specifies the default port to use if the PORT environment
// variable is not set. By default, Port will be set to 8080 if the environment
// variable is not set. If multiple ports are provided, the first non-zero port
// will be used.
func WithDefaultPort(ports ...uint16) ServiceLoadOption {
	return func(o *Service) {
		for _, port := range ports {
			if port != 0 {
				o.Port = port
				break
			}
		}
	}
}

// WithDefaultPortString specifies the default port to use if the PORT
// environment variable is not set. By default, Port will be set to 8080 if the
// environment variable is not set. If multiple ports are provided, the first
// non-zero port will be used.
func WithDefaultPortString(portStrings ...string) ServiceLoadOption {
	return func(o *Service) {
		for _, portString := range portStrings {
			if portString != "" {
				port, err := strconv.ParseUint(portString, 10, 16)
				if err == nil && port != 0 {
					o.Port = uint16(port)
					break
				}
			}
		}
	}
}

// WithDefaultServiceName specifies the default name to use if the K_SERVICE
// environment variable is not set. If multiple names are provided, the first
// non-empty name will be used.
func WithDefaultServiceName(names ...string) ServiceLoadOption {
	return func(o *Service) {
		for _, name := range names {
			if name != "" {
				o.Name = name
				break
			}
		}
	}
}

// WithDefaultRevision specifies the default revision to use if the K_REVISION
// environment variable is not set. If multiple revisions are provided, the
// first non-empty revision will be used.
func WithDefaultRevision(revisions ...string) ServiceLoadOption {
	return func(o *Service) {
		for _, revision := range revisions {
			if revision != "" {
				o.Revision = revision
				break
			}
		}
	}
}

// WithDefaultConfiguration specifies the default configuration to use if the
// K_CONFIGURATION environment variable is not set. If multiple configurations
// are provided, the first non-empty configuration will be used.
func WithDefaultConfiguration(configurations ...string) ServiceLoadOption {
	return func(o *Service) {
		for _, configuration := range configurations {
			if configuration != "" {
				o.Configuration = configuration
				break
			}
		}
	}
}

// LoadService loads configuration for a Cloud Run service from environment
// variables. It returns a Service containing the loaded configuration or
// ErrEnvironmentProcess if environment variable processing fails and/or
// ErrInvalidPort if the PORT environment variable is set to 0. Use options to
// specify default values for the service.
func LoadService(opts ...ServiceLoadOption) (*Service, error) {
	// Default values
	s := defaultService()

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Reload configuration from the environment
	if err := s.Reload(); err != nil {
		return nil, err
	}

	return s, nil
}

// Reload reloads the configuration for a Cloud Run service from environment
// variables. It returns ErrEnvironmentProcess if environment variable
// processing fails and/or ErrInvalidPort if the PORT environment variable is
// set to 0. It does not overwrite values already set in the Service struct if
// they are not set in the environment.
func (s *Service) Reload() error {
	if name := os.Getenv("K_SERVICE"); name != "" {
		s.Name = name
	}

	if revision := os.Getenv("K_REVISION"); revision != "" {
		s.Revision = revision
	}

	if configuration := os.Getenv("K_CONFIGURATION"); configuration != "" {
		s.Configuration = configuration
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return errors.Join(ErrEnvironmentProcess, ErrInvalidPort, err)
		}
		if port == 0 {
			return fmt.Errorf("%w: PORT value cannot be 0", ErrInvalidPort)
		}
		s.Port = uint16(port)
	}

	return nil
}

// EnvDecode implements the [envconfig.DecoderCtx] interface from
// github.com/sethvargo/go-envconfig. This ensures that [envconfig.Process] will
// return errors derived from [ErrEnvironmentProcess] and [ErrInvalidPort] if
// the environment variables are invalid.
//
// The behavior of this function is equivalent to initializing a Service with
// the default values and then calling [Service.Reload]. However, values already
// set in the Service struct prior to calling this function are not overridden
// by the defaults, only by the reloaded values from the environment.
//
// [envconfig.DecoderCtx]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#DecoderCtx
// [envconfig.Process]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#Process
func (s *Service) EnvDecode(ctx context.Context, val string) error {
	defaults := defaultService()

	// Name, Revision and Configuration default values are empty by default and
	// can be skipped.

	if s.Port == 0 {
		s.Port = defaults.Port
	}

	return s.Reload()
}
