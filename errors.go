package runcfg

import "errors"

var (
	// ErrInvalidEnvironmentVariableType indicates an invalid environment
	// variable type. The only valid types for Env* variables are string and
	// []string.
	ErrInvalidEnvironmentVariableType = errors.New("invalid environment variable type")

	// ErrEnvironmentProcess indicates a failure while processing configuration
	// from environment variables
	ErrEnvironmentProcess = errors.New("failed to process configuration from environment variables")

	// ErrMetadataFetch indicates a failure while fetching metadata from the
	// metadata server
	ErrMetadataFetch = errors.New("failed to fetch metadata from server")
)
