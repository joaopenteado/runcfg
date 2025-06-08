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
- Supports setting log level via `LOG_LEVEL` environment variable
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
	"net/http"
	"os"

	"github.com/joaopenteado/runcfg/zerologcfg"
	"github.com/rs/zerolog"
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

	// Example HTTP handler using the Handler middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())
		log.Info().Msg("Handling request")
		w.Write([]byte("Hello!"))
	})

	// Create a new HTTP server with the logger middleware
	// Ensure the Handler middleware is added after any tracing middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: zerologcfg.Handler(logger)(mux),
	}

	logger.Info().Msgf("Server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
```

## HTTP Middleware

The package provides an HTTP middleware `Handler` that injects the `zerolog.Logger`
into the request's context. This is particularly useful for ensuring that log
messages within HTTP handlers automatically include trace information if available
in the request context.

```go
// logger is your configured zerolog.Logger instance
// myRouter is your http.Handler (e.g., a chi router or http.ServeMux)
http.ListenAndServe(":8080", zerologcfg.Handler(logger)(myRouter))
```

It's important to install this middleware *after* any middleware that might
add tracing information to the request context (e.g., OpenTelemetry middleware).

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

### Setting Log Level via Environment Variable

You can set the global log level using the `LOG_LEVEL` environment variable.
The package will automatically configure zerolog with the specified level during
initialization.

Example:
```bash
export LOG_LEVEL=DEBUG
go run main.go
```

If the `LOG_LEVEL` environment variable is not set or contains an invalid value,
the package will use zerolog's default log level.

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
