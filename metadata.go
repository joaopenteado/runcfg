package runcfg

import (
	"context"
	"errors"
	"fmt"
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
	// EnvProjectID is a list of environment variable names that are used by
	// default to load the project ID. Variables are checked in order, with the
	// first non-empty value taking precedence. The values returned by the
	// metadata server will override these when MetadataProjectID is included
	// in the fields to fetch.
	EnvProjectID = []string{"CLOUDSDK_CORE_PROJECT", "GOOGLE_CLOUD_PROJECT_ID", "GCP_PROJECT_ID"}

	// EnvProjectNumber is a list of environment variable names that are used
	// by default to load the project number. Variables are checked in order,
	// with the first non-empty value taking precedence. The values returned by
	// the metadata server will override these when MetadataProjectNumber is
	// included in the fields to fetch.
	EnvProjectNumber = []string{"GOOGLE_CLOUD_PROJECT_NUMBER", "GCP_PROJECT_NUMBER"}

	// EnvRegion is a list of environment variable names that are used by
	// default to load the region. Variables are checked in order, with the
	// first non-empty value taking precedence. The values returned by the
	// metadata server will override these when MetadataRegion is included in
	// the fields to fetch.
	EnvRegion = []string{"CLOUDSDK_COMPUTE_REGION", "GOOGLE_CLOUD_REGION", "GCP_REGION"}

	// EnvInstanceID is a list of environment variable names that are used by
	// default to load the instance ID. Variables are checked in order, with the
	// first non-empty value taking precedence. The values returned by the
	// metadata server will override these when MetadataInstanceID is included
	// in the fields to fetch.
	EnvInstanceID = []string{"CLOUD_RUN_INSTANCE_ID"}

	// EnvServiceAccountEmail is a list of environment variable names that
	// are used by default to load the service account email. Variables are
	// checked in order, with the first non-empty value taking precedence. The
	// values returned by the metadata server will override these when
	// MetadataServiceAccountEmail is included in the fields to fetch.
	EnvServiceAccountEmail = []string{"GOOGLE_SERVICE_ACCOUNT_EMAIL"}
)

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

	// Unique identifier of the instance.
	InstanceID string

	// ServiceAccountEmail for the service identity of this Cloud Run service.
	ServiceAccountEmail string
}

func defaultMetadata() *Metadata {
	return &Metadata{
		ProjectID:           GetFirstEnv(EnvProjectID...),
		ProjectNumber:       GetFirstEnv(EnvProjectNumber...),
		Region:              GetFirstEnv(EnvRegion...),
		InstanceID:          GetFirstEnv(EnvInstanceID...),
		ServiceAccountEmail: GetFirstEnv(EnvServiceAccountEmail...),
	}
}

type MetadataLoadOption func(*Metadata)

// WithDefaultProjectID specifies the default project ID to use if the
// environment variable is not set. If multiple project IDs are provided, the
// first non-empty project ID will be used.
func WithDefaultProjectID(projectIDs ...string) MetadataLoadOption {
	return func(o *Metadata) {
		for _, projectID := range projectIDs {
			if projectID != "" {
				o.ProjectID = projectID
				break
			}
		}
	}
}

// WithDefaultProjectNumber specifies the default project number to use if the
// environment variable is not set. If multiple project numbers are provided,
// the first non-empty project number will be used.
func WithDefaultProjectNumber(projectNumbers ...string) MetadataLoadOption {
	return func(o *Metadata) {
		for _, projectNumber := range projectNumbers {
			if projectNumber != "" {
				o.ProjectNumber = projectNumber
				break
			}
		}
	}
}

// WithDefaultRegion specifies the default region to use if the environment
// variable is not set. If multiple regions are provided, the first non-empty
// region will be used.
func WithDefaultRegion(regions ...string) MetadataLoadOption {
	return func(o *Metadata) {
		for _, region := range regions {
			if region != "" {
				o.Region = region
				break
			}
		}
	}
}

// WithDefaultInstanceID specifies the default instance ID to use if the
// environment variable is not set. If multiple instance IDs are provided, the
// first non-empty instance ID will be used.
func WithDefaultInstanceID(instanceIDs ...string) MetadataLoadOption {
	return func(o *Metadata) {
		for _, instanceID := range instanceIDs {
			if instanceID != "" {
				o.InstanceID = instanceID
				break
			}
		}
	}
}

// WithDefaultServiceAccountEmail specifies the default service account email
// to use if the environment variable is not set. If multiple service account
// emails are provided, the first non-empty service account email will be used.
func WithDefaultServiceAccountEmail(serviceAccountEmails ...string) MetadataLoadOption {
	return func(o *Metadata) {
		for _, serviceAccountEmail := range serviceAccountEmails {
			if serviceAccountEmail != "" {
				o.ServiceAccountEmail = serviceAccountEmail
				break
			}
		}
	}
}

// LoadMetadata loads the metadata from the Cloud Run metadata server. It
// returns a pointer to a new Metadata struct with the loaded values.
//
// The metadataFields parameter controls which fields to fetch from
// the metadata server. If a field is not requested and not set via a default
// value, it will remain empty in the returned struct.
//
// By default, values not loaded from the metadata server will be loaded from
// the first non-empty value of the environment variables listed in
// EnvProjectID, EnvProjectNumber, EnvRegion, EnvInstanceID, and
// EnvServiceAccountEmail.
func LoadMetadata(ctx context.Context, metadataFields MetadataField, opts ...MetadataLoadOption) (*Metadata, error) {
	// Default values
	m := defaultMetadata()

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	// Reload metadata from the server
	if err := m.Reload(ctx, metadataFields); err != nil {
		return nil, err
	}

	return m, nil
}

// Reload reloads the metadata from the Cloud Run metadata server.
// The metadataFields parameter controls which fields to fetch from
// the metadata server. Fields not requested will not be loaded and won't
// overwrite values already set in the Metadata struct.
func (m *Metadata) Reload(ctx context.Context, metadataFields MetadataField) error {
	g, ctx := errgroup.WithContext(ctx)

	if metadataFields&MetadataProjectID != 0 {
		g.Go(func() error {
			projectID, err := metadata.ProjectIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch project ID: %w", err)
			}
			m.ProjectID = projectID
			return nil
		})
	}

	// If both project number and region are requested, both will be fetched
	// together from the instance/region metadata endpoint.
	if metadataFields&MetadataRegion != 0 {
		g.Go(func() error {
			res, err := metadata.GetWithContext(ctx, "instance/region")
			if err != nil {
				return fmt.Errorf("failed to fetch region: %w", err)
			}
			// Region is returned in the format projects/{num}/regions/{name}
			regionName := res[strings.LastIndexByte(res, '/')+1:]
			m.Region = regionName

			if metadataFields&MetadataProjectNumber != 0 {
				const projectNumberPrefixLen = len("projects/")
				projectNumber := res[projectNumberPrefixLen : strings.IndexByte(res[projectNumberPrefixLen:], '/')+projectNumberPrefixLen]
				m.ProjectNumber = projectNumber
			}
			return nil
		})
	} else if metadataFields&MetadataProjectNumber != 0 {
		g.Go(func() error {
			projectNumber, err := metadata.NumericProjectIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch project number: %w", err)
			}
			m.ProjectNumber = projectNumber
			return nil
		})
	}

	if metadataFields&MetadataInstanceID != 0 {
		g.Go(func() error {
			instanceID, err := metadata.InstanceIDWithContext(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch instance ID: %w", err)
			}
			m.InstanceID = instanceID
			return nil
		})
	}

	if metadataFields&MetadataServiceAccountEmail != 0 {
		g.Go(func() error {
			email, err := metadata.EmailWithContext(ctx, "default")
			if err != nil {
				return fmt.Errorf("failed to fetch service account email: %w", err)
			}
			m.ServiceAccountEmail = email
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Join(ErrMetadataFetch, err)
	}

	return nil
}

// EnvDecode implements the [envconfig.DecoderCtx] interface from
// github.com/sethvargo/go-envconfig. This ensures that [envconfig.Process] will
// return errors derived from [ErrMetadataFetch] if the metadata server is
// unavailable.
//
// The behavior of this function is equivalent to initializing a Metadata struct
// with the default values and then calling [Metadata.Reload] with values unset
// from the environment. Values already set in the Metadata struct prior to
// calling this function are not overridden by the defaults nor fetched from
// the metadata server.
//
// [envconfig.DecoderCtx]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#DecoderCtx
// [envconfig.Process]: https://pkg.go.dev/github.com/sethvargo/go-envconfig#Process
func (m *Metadata) EnvDecode(ctx context.Context, val string) error {
	metadataFields := MetadataNone
	defaults := defaultMetadata()

	if m.ProjectID == "" {
		if defaults.ProjectID != "" {
			m.ProjectID = defaults.ProjectID
		} else {
			metadataFields |= MetadataProjectID
		}
	}

	if m.ProjectNumber == "" {
		if defaults.ProjectNumber != "" {
			m.ProjectNumber = defaults.ProjectNumber
		} else {
			metadataFields |= MetadataProjectNumber
		}
	}

	if m.Region == "" {
		if defaults.Region != "" {
			m.Region = defaults.Region
		} else {
			metadataFields |= MetadataRegion
		}
	}

	if m.InstanceID == "" {
		if defaults.InstanceID != "" {
			m.InstanceID = defaults.InstanceID
		} else {
			metadataFields |= MetadataInstanceID
		}
	}

	if m.ServiceAccountEmail == "" {
		if defaults.ServiceAccountEmail != "" {
			m.ServiceAccountEmail = defaults.ServiceAccountEmail
		} else {
			metadataFields |= MetadataServiceAccountEmail
		}
	}

	return m.Reload(ctx, metadataFields)
}
