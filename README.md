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
    // Load service configuration with default options
    cfg, err := runcfg.LoadService()
    if err != nil {
        log.Fatal(err)
    }

    // Access configuration values
    fmt.Printf("Service name: %s\n", cfg.Name)
    fmt.Printf("Port: %d\n", cfg.Port)
    fmt.Printf("Revision: %s\n", cfg.Revision)
    fmt.Printf("Configuration: %s\n", cfg.Configuration)
}
```

```go
// With custom options
cfg, err := runcfg.LoadService(
    runcfg.WithDefaultPort(3000),
    runcfg.WithDefaultServiceName("my-service"),
    runcfg.WithDefaultRevision("v1"),
    runcfg.WithDefaultConfiguration("prod"),
)
```

### Cloud Run Job Configuration

```go
import "github.com/joaopenteado/runcfg"

func main() {
    // Load job configuration
    cfg, err := runcfg.LoadJob()
    if err != nil {
        log.Fatal(err)
    }

    // Access configuration values
    fmt.Printf("Job name: %s\n", cfg.Name)
    fmt.Printf("Execution: %s\n", cfg.Execution)
    fmt.Printf("Task index: %d\n", cfg.TaskIndex)
    fmt.Printf("Task attempt: %d\n", cfg.TaskAttempt)
    fmt.Printf("Task count: %d\n", cfg.TaskCount)
}
```

```go
// With custom options
cfg, err := runcfg.LoadJob(
    runcfg.WithDefaultJobName("my-job"),
    runcfg.WithDefaultExecution("exec-1"),
    runcfg.WithDefaultTaskIndex(0),
    runcfg.WithDefaultTaskAttempt(0),
    runcfg.WithDefaultTaskCount(1),
)
```

### Metadata Configuration

```go
import "github.com/joaopenteado/runcfg"

func main() {
    ctx := context.Background()

    // Load metadata with specific fields
    cfg, err := runcfg.LoadMetadata(ctx, runcfg.MetadataProjectID | runcfg.MetadataRegion)
    if err != nil {
        log.Fatal(err)
    }

    // Access metadata values
    fmt.Printf("Project ID: %s\n", cfg.ProjectID)
    fmt.Printf("Region: %s\n", cfg.Region)
}
```

```go
// With custom options
cfg, err := runcfg.LoadMetadata(ctx, runcfg.MetadataAll,
    runcfg.WithDefaultProjectID("my-project"),
    runcfg.WithDefaultRegion("us-central1"),
    runcfg.WithDefaultProjectNumber("123456789"),
    runcfg.WithDefaultInstanceID("instance-1"),
    runcfg.WithDefaultServiceAccountEmail("service-account@project.iam.gserviceaccount.com"),
)
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
`Env*` package variables. Each variable can be set to a list of fallback
variables ([]string) that will be checked in order until a non-empty value is
found. If no non-empty environment variables are found and the loader is
configured to fetch the corresponding metadata field, the value will be fetched
from the metadata server.

## Error Handling

The package defines the following error types:

```go
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
```

## License

MIT
