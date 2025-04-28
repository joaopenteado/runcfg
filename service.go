package runcfg

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type serviceLoadOptions struct {
	defaultPort          uint16
	defaultName          string
	defaultRevision      string
	defaultConfiguration string
}

func defaultServiceLoadOptions() serviceLoadOptions {
	return serviceLoadOptions{
		defaultPort: 8080,
	}
}

type ServiceLoadOption func(*serviceLoadOptions)

// WithDefaultPort specifies the default port to use if the PORT environment
// variable is not set. By default, Port will be set to 8080 if the environment
// variable is not set. If multiple ports are provided, the first non-zero port
// will be used.
func WithDefaultPort(ports ...uint16) ServiceLoadOption {
	return func(o *serviceLoadOptions) {
		for _, port := range ports {
			if port != 0 {
				o.defaultPort = port
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
	return func(o *serviceLoadOptions) {
		for _, portString := range portStrings {
			if portString != "" {
				port, err := strconv.ParseUint(portString, 10, 16)
				if err == nil && port != 0 {
					o.defaultPort = uint16(port)
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
	return func(o *serviceLoadOptions) {
		for _, name := range names {
			if name != "" {
				o.defaultName = name
				break
			}
		}
	}
}

// WithDefaultRevision specifies the default revision to use if the K_REVISION
// environment variable is not set. If multiple revisions are provided, the
// first non-empty revision will be used.
func WithDefaultRevision(revisions ...string) ServiceLoadOption {
	return func(o *serviceLoadOptions) {
		for _, revision := range revisions {
			if revision != "" {
				o.defaultRevision = revision
				break
			}
		}
	}
}

// WithDefaultConfiguration specifies the default configuration to use if the
// K_CONFIGURATION environment variable is not set. If multiple configurations
// are provided, the first non-empty configuration will be used.
func WithDefaultConfiguration(configurations ...string) ServiceLoadOption {
	return func(o *serviceLoadOptions) {
		for _, configuration := range configurations {
			if configuration != "" {
				o.defaultConfiguration = configuration
				break
			}
		}
	}
}

// Service contains environment variables available to Cloud Run services.
// All variables except PORT are available to all containers. The PORT variable
// is only added to the ingress container. For more details see:
// https://cloud.google.com/run/docs/container-contract#services-env-vars
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

// LoadService loads configuration for a Cloud Run service from environment
// variables. It returns a Service containing the loaded configuration or
// ErrEnvironmentProcess if environment variable processing fails and/or
// ErrInvalidPort if the PORT environment variable is set to 0.
func LoadService(opts ...ServiceLoadOption) (*Service, error) {
	s := &Service{}
	return s, s.Reload(opts...)
}

// Reload reloads the configuration for a Cloud Run service from environment
// variables. It returns ErrEnvironmentProcess if environment variable
// processing fails and/or ErrInvalidPort if the PORT environment variable is
// set to 0.
func (s *Service) Reload(opts ...ServiceLoadOption) error {
	loadOpts := defaultServiceLoadOptions()
	for _, opt := range opts {
		opt(&loadOpts)
	}

	s.Name = os.Getenv("K_SERVICE")
	s.Revision = os.Getenv("K_REVISION")
	s.Configuration = os.Getenv("K_CONFIGURATION")

	if s.Name == "" {
		s.Name = loadOpts.defaultName
	}
	if s.Revision == "" {
		s.Revision = loadOpts.defaultRevision
	}
	if s.Configuration == "" {
		s.Configuration = loadOpts.defaultConfiguration
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return errors.Join(ErrEnvironmentProcess, ErrInvalidPort, err)
		}
		s.Port = uint16(port)
	} else {
		s.Port = loadOpts.defaultPort
	}

	if s.Port == 0 {
		return fmt.Errorf("%w: PORT value cannot be 0", ErrInvalidPort)
	}

	return nil
}
