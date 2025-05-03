# zerologcfg

[![Go Reference](https://pkg.go.dev/badge/github.com/joaopenteado/runcfg/zerologcfg.svg)](https://pkg.go.dev/github.com/joaopenteado/runcfg/zerologcfg)

A zerolog configuration package for Google Cloud Run that provides proper
integration with Cloud Logging.

## Acknowledgements

Heavily inspired by the excellent
[yfuruyama/crzerolog](https://github.com/yfuruyama/crzerolog), with some minor
tweaks for my own personal use.

## Features

- Configures zerolog to match Cloud Logging's severity levels and timestamp
format
- Automatically adds source location information (file, line, function)
- Integrates with OpenTelemetry tracing (trace ID, span ID, and sampling status)
- Formats logs according to Cloud Logging's structured logging requirements
- Supports Go 1.24.1 and above

## Installation

```bash
go get github.com/joaopenteado/runcfg/zerologcfg
```

## Usage

```go
package main

import (
    "context"
    "os"
    "github.com/rs/zerolog"
    "github.com/joaopenteado/runcfg/zerologcfg"
)

func main() {
    // Create a new logger with Cloud Run configuration
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

    // Add Cloud Logging hook with your project ID
    logger = logger.Hook(zerologcfg.Hook("your-project-id"))

    // Use the logger
    logger.Info().Msg("Hello from Cloud Run!")

    // Log with context containing trace information
    ctx := context.Background() // Your context with trace info
    logger.Info().Ctx(ctx).Msg("Log with trace information")
}
```

## Log Levels

The package maps zerolog levels to Cloud Logging severity levels:

- TRACE → DEFAULT
- DEBUG → DEBUG
- INFO → INFO
- WARN → WARNING
- ERROR → ERROR
- FATAL → CRITICAL
- PANIC → ALERT
- NoLevel → DEFAULT

## Structured Logging

The package automatically adds the following fields to your logs:

- `severity`: The log level
- `time`: Timestamp in RFC3339Nano format
- `logging.googleapis.com/sourceLocation`: Contains file, line, and function
information
- `logging.googleapis.com/trace`: Trace ID (when available)
- `logging.googleapis.com/spanId`: Span ID (when available)
- `logging.googleapis.com/trace_sampled`: Trace sampling status (when available)

## OpenTelemetry Integration

The package automatically integrates with OpenTelemetry tracing when a valid
span context is present in the logging context. It adds the following fields:

- Trace ID in the format `projects/{project-id}/traces/{trace-id}`
- Span ID
- Trace sampling status (true/false)

## License

MIT
