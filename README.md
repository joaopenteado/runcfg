# runcfg

[![Go Reference](https://pkg.go.dev/badge/github.com/joaopenteado/runcfg.svg)](https://pkg.go.dev/github.com/joaopenteado/runcfg)

A Go package for loading configuration in Google Cloud Run services and jobs.
This package provides a simple and type-safe way to access environment variables
and metadata server information in Cloud Run environments.

## Features

- Load configuration for both Cloud Run services and jobs
- Access environment variables in a type-safe way
- Fetch metadata from the Cloud Run metadata server
- Support for environment variable overrides
- Concurrent metadata fetching for better performance
- Configurable metadata field loading

## Installation

```bash
go get -u github.com/joaopenteado/runcfg
```

## Usage

### Cloud Run Service Configuration

```go
import "github.com/joaopenteado/runcfg"

func main() {
    ctx := context.Background()

    // Load service configuration with default options
    cfg, err := runcfg.LoadService(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Access configuration values
    fmt.Printf("Service name: %s\n", cfg.Name)
    fmt.Printf("Port: %d\n", cfg.Port)
    fmt.Printf("Project ID: %s\n", cfg.ProjectID)
}
```

```go
// With custom options
cfg, err := runcfg.LoadService(ctx,
    runcfg.WithMetadata(runcfg.MetadataProjectID | runcfg.MetadataRegion),
    runcfg.WithDefaultPort(3000),
)
```

### Cloud Run Job Configuration

```go
import "github.com/joaopenteado/runcfg"

func main() {
    ctx := context.Background()

    // Load job configuration
    cfg, err := runcfg.LoadJob(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Access configuration values
    fmt.Printf("Job name: %s\n", cfg.Name)
    fmt.Printf("Task index: %d\n", cfg.TaskIndex)
    fmt.Printf("Project ID: %s\n", cfg.ProjectID)
}
```

## Configuration Options

### Metadata Fields

The package supports loading different metadata fields:

```go
const (
    MetadataNone MetadataField = 0
    MetadataProjectID MetadataField = 1 << iota
    MetadataProjectNumber
    MetadataRegion
    MetadataInstanceID
    MetadataServiceAccountEmail
    MetadataAll = ^MetadataField(0)
)
```

### Environment Variables Overrides for Metadata Information

The package allows overriding metadata values using environment variables. When
an environment variable is set, its value takes precedence and the corresponding
metadata field will not be fetched from the metadata server.

The following environment variables are supported by default:

- Project ID: Checks in order: `CLOUDSDK_CORE_PROJECT`, `GOOGLE_CLOUD_PROJECT`, `GCP_PROJECT_ID`
- Project Number: `GCP_PROJECT_NUMBER`
- Region: Checks in order: `CLOUDSDK_COMPUTE_REGION`, `GCP_REGION`
- Instance ID: `CLOUD_RUN_INSTANCE_ID`
- Service Account Email: `GOOGLE_SERVICE_ACCOUNT_EMAIL`

You can customize which environment variables are checked by modifying the
`Env*` package variables. Each variable can be set to either a single
environment variable name (string) or a list of fallback variables ([]string)
that will be checked in order until a non-empty value is found. If no non-empty
environment variables are found and the loader is configured to fetch the
corresponding metadata field, the value will be fetched from the metadata
server.


## Error Handling

The package defines the following types, from which all the errors it returns
are based on.

```go
var (
    ErrInvalidEnvironmentVariableType = errors.New("invalid environment variable type")
    ErrEnvironmentProcess = errors.New("failed to process configuration from environment variables")
    ErrMetadataFetch = errors.New("failed to fetch metadata from server")
)
```

## License

MIT
