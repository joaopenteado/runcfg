package runcfg

import "errors"

var (
	// ErrEnvironmentProcess indicates a failure while processing configuration
	// from environment variables.
	ErrEnvironmentProcess = errors.New("failed to process configuration from environment variables")

	// ErrInvalidPort indicates an invalid port value was specified.
	// The PORT environment variable must be a valid port number, ranging from
	// 1 to 65535.
	ErrInvalidPort = errors.New("invalid PORT value")

	// ErrMetadataFetch indicates a failure while fetching metadata from the
	// metadata server.
	ErrMetadataFetch = errors.New("failed to fetch metadata from server")
)
