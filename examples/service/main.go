package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joaopenteado/runcfg"
	"github.com/joaopenteado/runcfg/otelcfg"
	"github.com/joaopenteado/runcfg/zerologcfg"
	"github.com/riandyrn/otelchi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/sethvargo/go-envconfig"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

const (
	ShutdownTimeout = 10 * time.Second
	ServerTimeout   = 10 * time.Second
)

var (
	ErrInitialization = errors.New("initialization failed")
	ErrShutdown       = errors.New("graceful shutdown failed")
)

type config struct {
	runcfg.Metadata `env:"RUNCFG_METADATA, decodeunset"`
	runcfg.Service  `env:"RUNCFG_SERVICE, decodeunset"`
	Greeting        string `env:"GREETING, default=Hello"`
}

func main() {
	// Logging and error handling
	// Ensure a proper error exit code is returned in the case of an error
	var returnErr error
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	defer func() {
		if returnErr != nil {
			if errors.Is(returnErr, ErrInitialization) || errors.Is(returnErr, ErrShutdown) {
				logger.Fatal().Err(errors.Unwrap(returnErr)).Msg(returnErr.Error())
				return
			}

			logger.Fatal().Err(returnErr).Msg("unknown return error")
		}
	}()

	// Base context and graceful shutdown context
	// The cancel function should not be used, its sole purpose is to signal
	// that the main function is done.
	// shutdownCtx blocks until the graceful shutdown signal is received.
	// shutdown starts the graceful shutdown process, with a timeout of 10s
	// https://cloud.google.com/run/docs/samples/cloudrun-sigterm-handler
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCtx, shutdown := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	shutdownDone := make(chan struct{})
	defer func() {
		cancel()       // Signals that the main function is done before
		shutdown()     // Ensure this context is properly cleaned up
		<-shutdownDone // Waits for shutdown goroutine to finish
	}()
	go func() {
		defer close(shutdownDone)
		<-shutdownCtx.Done()

		// Check if the main function is already done for an early return
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Start shutdown timer
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, ShutdownTimeout)
		defer timeoutCancel()

		// Wait for either the main function to finish or the timeout
		<-timeoutCtx.Done()
		cancel() // Forcibly cancel the base context if necessary
	}()

	// Load configuration
	cfg := &config{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		returnErr = errors.Join(ErrInitialization, err)
		return
	}

	// Setup logger for Cloud Logging
	logger = logger.Hook(zerologcfg.Hook(cfg.ProjectID))

	// Initialize tracing
	otelShutdown, err := setupOpenTelemetry(ctx, cfg)
	if err != nil {
		returnErr = errors.Join(ErrInitialization, err)
		return
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			returnErr = errors.Join(ErrShutdown, err)
		}
	}()

	// Setup Cloud Profier
	profCfg := profiler.Config{
		Service:        cfg.Name,
		ServiceVersion: cfg.Revision,
		ProjectID:      cfg.ProjectID,
		Instance:       cfg.InstanceID,
		Zone:           cfg.Region,
	}
	if err := profiler.Start(profCfg); err != nil {
		returnErr = errors.Join(ErrInitialization, err)
		return
	}

	// Setup HTTP server
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Timeout(ServerTimeout),
		middleware.Heartbeat("/healthz"), // Liveness probe
		hlog.NewHandler(logger),
		otelchi.Middleware(cfg.Name, otelchi.WithChiRoutes(r)),
	)

	srv := &http.Server{
		Addr:    ":" + strconv.FormatUint(uint64(cfg.Port), 10),
		Handler: r,
	}

	go func() {
		if srvErr := srv.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			returnErr = errors.Join(ErrInitialization, srvErr)
			shutdown() // initiate a graceful shutdown
		}
	}()

	// Wait for the graceful shutdown signal
	<-shutdownCtx.Done()
	// The base context will be forcibly cancelled in 10s from now

	g, gCtx := errgroup.WithContext(ctx)

	// Shutdown the server
	g.Go(func() error {
		if err := srv.Shutdown(gCtx); err != nil {
			// Invidually log the error, since g.Wait() will only return the
			// first error
			logger.Err(err).Msg("failed to gracefully shutdown server")
			return err
		}
		return nil
	})

	if gErr := g.Wait(); gErr != nil {
		// Ensures a non-sucessful exit code is returned
		returnErr = errors.Join(ErrShutdown, gErr)
	}
}

func setupOpenTelemetry(ctx context.Context, cfg *config) (func(context.Context) error, error) {
	svc := otelcfg.NewServiceResource(&cfg.Metadata, &cfg.Service, otelcfg.WithAttributes(
		// Required to assign a project to the traces in Cloud Trace
		// https://cloud.google.com/trace/docs/migrate-to-otlp-endpoints
		otelattr.String("gcp.project_id", cfg.ProjectID),
	))

	res, err := resource.Merge(resource.Default(), svc)
	if err != nil {
		return nil, err
	}

	return otelcfg.SetupTracerProvider(ctx,
		otelcfg.WithTracerProviderOptions(trace.WithResource(res)),
	)
}
