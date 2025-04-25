package runcfg

import "errors"

var (
	// ErrEnvironmentProcess indicates a failure while processing configuration
	// from environment variables
	ErrEnvironmentProcess = errors.New("failed to process configuration from environment variables")

	// ErrMetadataFetch indicates a failure while fetching metadata from the
	// metadata server
	ErrMetadataFetch = errors.New("failed to fetch metadata from server")
)
