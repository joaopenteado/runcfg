package runcfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/sync/errgroup"
)

// MetadataField represents a specific metadata field that can be loaded.
type MetadataField uint

const (
	// MetadataNone represents no metadata fields.
	MetadataNone MetadataField = 0

	// MetadataProjectID represents the project ID.
	MetadataProjectID MetadataField = 1 << iota

	// MetadataProjectNumber represents the project number.
	MetadataProjectNumber

	// MetadataRegion represents the region.
	MetadataRegion

	// MetadataInstanceID represents the instance ID.
	MetadataInstanceID

	// MetadataServiceAccountEmail represents the service account email.
	MetadataServiceAccountEmail

	// MetadataAll represents all metadata fields.
	MetadataAll = ^MetadataField(0)
)

var (
	// EnvProjectID is a list of environment variable names that can override
	// the project ID from the metadata server. Variables are checked in order,
	// with the first non-empty value taking precedence.
	EnvProjectID = []string{"CLOUDSDK_CORE_PROJECT", "GOOGLE_CLOUD_PROJECT", "GCP_PROJECT_ID"}

	// EnvProjectNumber is a list of environment variable names that can
	// override the project number from the metadata server. Variables are
	// checked in order, with the first non-empty value taking precedence.
	EnvProjectNumber = []string{"GCP_PROJECT_NUMBER"}

	// EnvRegion is a list of environment variable names that can override
	// the region from the metadata server. Variables are checked in order,
	// with the first non-empty value taking precedence.
	EnvRegion = []string{"CLOUDSDK_COMPUTE_REGION", "GCP_REGION"}

	// EnvInstanceID is a list of environment variable names that can override
	// the instance ID from the metadata server. Variables are checked in order,
	// with the first non-empty value taking precedence.
	EnvInstanceID = []string{"CLOUD_RUN_INSTANCE_ID"}

	// EnvServiceAccountEmail is a list of environment variable names that can
	// override the service account email from the metadata server. Variables
	// are checked in order, with the first non-empty value taking precedence.
	EnvServiceAccountEmail = []string{"GOOGLE_SERVICE_ACCOUNT_EMAIL"}
)

// getEnv retrieves environment variable values from a list of environment
// variable names. It checks each environment variable in order and returns the
// first non-empty value found. If no value is found, returns an empty string.
func getEnv(key []string) string {
	for _, v := range key {
		if val := os.Getenv(v); val != "" {
			return val
		}
	}

	return ""
}

// Metadata contains information from the instance metadata server.
// Cloud Run instances expose a metadata server that provides details about
// containers, such as project ID, region, instance ID, and service accounts.
// These values can be overridden by environment variables.
// For more details see:
// https://cloud.google.com/run/docs/container-contract#metadata-server
type Metadata struct {
	// Project ID of the project the Cloud Run service belongs to.
	ProjectID string

	// Project number of the project the Cloud Run service belongs to.
	ProjectNumber string

	// Region of this Cloud Run service.
	Region string

	// Unique identifier of the instance (also available in logs).
	InstanceID string

	// ServiceAccountEmail for the service identity of this Cloud Run service.
	ServiceAccountEmail string
}

// LoadMetadata retrieves metadata information from the Cloud Run metadata
// server and environment variables. Environment variables take precedence over
// metadata server values. The metadataFields parameter controls which fields
// to fetch from the metadata server - if a field is not requested and not set
// via environment variables, it will remain empty in the returned struct.
func LoadMetadata(ctx context.Context, metadataFields MetadataField) (*Metadata, error) {
	cfg := Metadata{
		ProjectID:           getEnv(EnvProjectID),
		ProjectNumber:       getEnv(EnvProjectNumber),
		Region:              getEnv(EnvRegion),
		InstanceID:          getEnv(EnvInstanceID),
		ServiceAccountEmail: getEnv(EnvServiceAccountEmail),
	}

	g, ctx := errgroup.WithContext(ctx)

	if cfg.ProjectID == "" && metadataFields&MetadataProjectID != 0 {
		g.Go(func() error {
			projectID, err := metadata.ProjectIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch project ID: %w", err)
			}
			cfg.ProjectID = projectID
			return nil
		})
	}

	// If both project number and region are requested, both will be fetched
	// together from the instance/region metadata endpoint.
	if cfg.Region == "" && metadataFields&MetadataRegion != 0 {
		g.Go(func() error {
			res, err := metadata.GetWithContext(ctx, "instance/region")
			if err != nil {
				return fmt.Errorf("failed to fetch region: %w", err)
			}
			// Region is returned in the format projects/{num}/regions/{name}
			regionName := res[strings.LastIndexByte(res, '/')+1:]
			cfg.Region = regionName

			if cfg.ProjectNumber == "" && metadataFields&MetadataProjectNumber != 0 {
				const projectNumberPrefixLen = len("projects/")
				projectNumber := res[projectNumberPrefixLen : strings.IndexByte(res[projectNumberPrefixLen:], '/')+projectNumberPrefixLen]
				cfg.ProjectNumber = projectNumber
			}
			return nil
		})
	} else if cfg.ProjectNumber == "" && metadataFields&MetadataProjectNumber != 0 {
		g.Go(func() error {
			projectNumber, err := metadata.NumericProjectIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch project number: %w", err)
			}
			cfg.ProjectNumber = projectNumber
			return nil
		})
	}

	if cfg.InstanceID == "" && metadataFields&MetadataInstanceID != 0 {
		g.Go(func() error {
			instanceID, err := metadata.InstanceIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch instance ID: %w", err)
			}
			cfg.InstanceID = instanceID
			return nil
		})
	}

	if cfg.ServiceAccountEmail == "" && metadataFields&MetadataServiceAccountEmail != 0 {
		g.Go(func() error {
			email, err := metadata.EmailWithContext(ctx, "default")
			if err != nil {
				return fmt.Errorf("failed to fetch service account email: %w", err)
			}
			cfg.ServiceAccountEmail = email
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, errors.Join(ErrMetadataFetch, err)
	}

	return &cfg, nil
}
